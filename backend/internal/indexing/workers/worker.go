package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	spotnet_client "github.com/superindexer/indexer/internal/spotnet/client"
	"github.com/superindexer/indexer/internal/spotnet/metadata"
	"github.com/superindexer/indexer/internal/spotnet/parser"
	usenet_client "github.com/superindexer/indexer/internal/usenet/client"
	"github.com/superindexer/indexer/internal/usenet/connection"
)

// Worker represents an indexing worker that processes jobs
type Worker struct {
	ID                int64
	SourceRepo        repositories.SpotnetSourceRepository
	PostRepo          repositories.SpotnetPostRepository
	ArticleRepo       repositories.ArticleRepository
	IndexingJobRepo   repositories.IndexingJobRepository
	NewsgroupRepo     repositories.NewsgroupRepository
	SpotnetClient     *spotnet_client.SpotnetClient
	UsenetClient      *usenet_client.NNTPClient
	PostParser        *parser.PostParser
	MetadataProcessor *metadata.MetadataProcessor
	ConnectionManager *connection.ConnectionManager
	log               *slog.Logger
}

// NewWorker creates a new indexing worker
func NewWorker(
	id int64,
	sourceRepo repositories.SpotnetSourceRepository,
	postRepo repositories.SpotnetPostRepository,
	articleRepo repositories.ArticleRepository,
	indexingJobRepo repositories.IndexingJobRepository,
	newsgroupRepo repositories.NewsgroupRepository,
	spotnetClient *spotnet_client.SpotnetClient,
	usenetClient *usenet_client.NNTPClient,
	postParser *parser.PostParser,
	metadataProcessor *metadata.MetadataProcessor,
) *Worker {
	return &Worker{
		ID:                id,
		SourceRepo:        sourceRepo,
		PostRepo:          postRepo,
		ArticleRepo:       articleRepo,
		IndexingJobRepo:   indexingJobRepo,
		NewsgroupRepo:     newsgroupRepo,
		SpotnetClient:     spotnetClient,
		UsenetClient:      usenetClient,
		PostParser:        postParser,
		MetadataProcessor: metadataProcessor,
		log:               slog.Default(),
	}
}

// ProcessJob dispatches to the appropriate processor based on job type
func (w *Worker) ProcessJob(ctx context.Context, job *models.IndexingJob) error {
	switch job.JobType {
	case "spotnet":
		return w.ProcessSpotnetJob(ctx, job)
	case "nntp":
		return w.ProcessUsenetJob(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}
}

// ProcessSpotnetJob processes a Spotnet indexing job with retry logic
func (w *Worker) ProcessSpotnetJob(ctx context.Context, job *models.IndexingJob) error {
	// Get source ID
	var sourceID int64
	if job.SourceID != nil {
		sourceID = *job.SourceID
	} else {
		return fmt.Errorf("spotnet job requires a source ID")
	}

	// Get source from database
	source, err := w.SourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source %d: %w", sourceID, err)
	}

	if !source.Enabled {
		return fmt.Errorf("source %d is disabled", sourceID)
	}

	// Update job status to running
	now := time.Now().UTC()
	job.Status = "running"
	job.StartedAt = &now
	if err := w.IndexingJobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status to running: %w", err)
	}

	// Determine last sync time for incremental fetching
	var sinceDate time.Time
	if source.LastSyncAt != nil {
		sinceDate = *source.LastSyncAt
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation
		if ctx.Err() != nil {
			w.markJobFailed(job, ctx.Err().Error())
			return ctx.Err()
		}

		w.log.Info("Fetching posts from source",
			"worker_id", w.ID,
			"source_id", source.ID,
			"source_name", source.Name,
			"attempt", attempt+1,
			"max_retries", maxRetries,
		)

		// Fetch new posts using the SpotnetClient with sinceDate filter
		posts, err := w.SpotnetClient.GetPosts(ctx, source, "", sinceDate)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch posts from source %d: %w", sourceID, err)
			w.log.Error("Failed to fetch posts",
				"source_id", sourceID,
				"attempt", attempt+1,
				"error", err,
			)

			// Calculate backoff delay: 1s, 2s, 4s...
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				w.log.Info("Retrying after backoff",
					"source_id", sourceID,
					"attempt", attempt+1,
					"backoff_seconds", backoff.Seconds(),
				)
				select {
				case <-ctx.Done():
					w.markJobFailed(job, ctx.Err().Error())
					return ctx.Err()
				case <-time.After(backoff):
					// Continue to next attempt
				}
			}
			continue
		}

		// Process and store posts
		processed, failed, err := w.storePosts(ctx, job, posts)
		if err != nil {
			lastErr = fmt.Errorf("failed to store posts: %w", err)
			w.log.Error("Failed to store posts",
				"source_id", sourceID,
				"attempt", attempt+1,
				"error", err,
			)

			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				select {
				case <-ctx.Done():
					w.markJobFailed(job, ctx.Err().Error())
					return ctx.Err()
				case <-time.After(backoff):
					// Continue to next attempt
				}
			}
			continue
		}

		// Update source last sync time
		updatedAt := time.Now().UTC()
		source.LastSyncAt = &updatedAt
		if err := w.SourceRepo.Update(ctx, source); err != nil {
			w.log.Error("Failed to update source last sync time",
				"source_id", sourceID,
				"error", err,
			)
		}

		// Update job progress
		progress := 100.0
		if len(posts) > 0 {
			progress = float64(processed) / float64(len(posts)) * 100.0
		}
		if err := w.IndexingJobRepo.UpdateProgress(ctx, job.ID, progress, processed, failed); err != nil {
			w.log.Error("Failed to update job progress", "job_id", job.ID, "error", err)
		}

		// Mark job as completed
		completedAt := time.Now().UTC()
		job.Status = "completed"
		job.RecordsProcessed = processed
		job.RecordsFailed = failed
		job.Progress = progress
		job.CompletedAt = &completedAt
		if err := w.IndexingJobRepo.Complete(ctx, job.ID); err != nil {
			w.log.Error("Failed to complete job", "job_id", job.ID, "error", err)
		}

		w.log.Info("Spotnet job completed successfully",
			"job_id", job.ID,
			"source_id", sourceID,
			"processed", processed,
			"failed", failed,
		)

		return nil
	}

	// All retries exhausted
	w.markJobFailed(job, fmt.Sprintf("all %d retries exhausted: %v", maxRetries, lastErr))
	return fmt.Errorf("spotnet job failed after %d retries: %w", maxRetries, lastErr)
}

// ProcessUsenetJob processes a Usenet (NNTP) indexing job with retry logic
func (w *Worker) ProcessUsenetJob(ctx context.Context, job *models.IndexingJob) error {
	// Get newsgroup ID
	var newsgroupID int64
	if job.NewsgroupID != nil {
		newsgroupID = *job.NewsgroupID
	} else {
		return fmt.Errorf("nntp job requires a newsgroup ID")
	}

	// Get newsgroup info
	newsgroup, err := w.NewsgroupRepo.GetByID(ctx, newsgroupID)
	if err != nil {
		return fmt.Errorf("failed to get newsgroup %d: %w", newsgroupID, err)
	}

	if !newsgroup.Enabled {
		return fmt.Errorf("newsgroup %d is disabled", newsgroupID)
	}

	// Get NNTP server (use the first enabled server)
	servers, err := w.getNNTPServers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get NNTP servers: %w", err)
	}

	if len(servers) == 0 {
		return fmt.Errorf("no enabled NNTP servers configured")
	}

	// Update job status to running
	now := time.Now().UTC()
	job.Status = "running"
	job.StartedAt = &now
	if err := w.IndexingJobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status to running: %w", err)
	}

	// Determine last indexed article number
	var lastArticleNum int64
	if newsgroup.LastIndexedArticleNum != nil {
		lastArticleNum = *newsgroup.LastIndexedArticleNum
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation
		if ctx.Err() != nil {
			w.markJobFailed(job, ctx.Err().Error())
			return ctx.Err()
		}

		w.log.Info("Fetching articles from newsgroup",
			"worker_id", w.ID,
			"newsgroup_id", newsgroupID,
			"newsgroup_name", newsgroup.Name,
			"attempt", attempt+1,
			"max_retries", maxRetries,
		)

		// Fetch new articles using NNTP
		articles, err := w.fetchArticles(ctx, newsgroup, lastArticleNum)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch articles from newsgroup %d: %w", newsgroupID, err)
			w.log.Error("Failed to fetch articles",
				"newsgroup_id", newsgroupID,
				"attempt", attempt+1,
				"error", err,
			)

			// Calculate backoff delay: 1s, 2s, 4s...
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				w.log.Info("Retrying after backoff",
					"newsgroup_id", newsgroupID,
					"attempt", attempt+1,
					"backoff_seconds", backoff.Seconds(),
				)
				select {
				case <-ctx.Done():
					w.markJobFailed(job, ctx.Err().Error())
					return ctx.Err()
				case <-time.After(backoff):
					// Continue to next attempt
				}
			}
			continue
		}

		// Process and store articles
		processed, failed, err := w.storeArticles(ctx, job, newsgroup, articles)
		if err != nil {
			lastErr = fmt.Errorf("failed to store articles: %w", err)
			w.log.Error("Failed to store articles",
				"newsgroup_id", newsgroupID,
				"attempt", attempt+1,
				"error", err,
			)

			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				select {
				case <-ctx.Done():
					w.markJobFailed(job, ctx.Err().Error())
					return ctx.Err()
				case <-time.After(backoff):
					// Continue to next attempt
				}
			}
			continue
		}

		// Update newsgroup last indexed info
		if len(articles) > 0 {
			updatedAt := time.Now().UTC()
			if err := w.NewsgroupRepo.UpdateLastIndexed(ctx, newsgroup.ID, articles[len(articles)-1].ID, updatedAt); err != nil {
				w.log.Error("Failed to update newsgroup last indexed",
					"newsgroup_id", newsgroupID,
					"error", err,
				)
			}
		}

		// Update job progress
		progress := 100.0
		if len(articles) > 0 {
			progress = float64(processed) / float64(len(articles)) * 100.0
		}
		if err := w.IndexingJobRepo.UpdateProgress(ctx, job.ID, progress, processed, failed); err != nil {
			w.log.Error("Failed to update job progress", "job_id", job.ID, "error", err)
		}

		// Mark job as completed
		completedAt := time.Now().UTC()
		job.Status = "completed"
		job.RecordsProcessed = processed
		job.RecordsFailed = failed
		job.Progress = progress
		job.CompletedAt = &completedAt
		if err := w.IndexingJobRepo.Complete(ctx, job.ID); err != nil {
			w.log.Error("Failed to complete job", "job_id", job.ID, "error", err)
		}

		w.log.Info("NNTP job completed successfully",
			"job_id", job.ID,
			"newsgroup_id", newsgroupID,
			"processed", processed,
			"failed", failed,
		)

		return nil
	}

	// All retries exhausted
	w.markJobFailed(job, fmt.Sprintf("all %d retries exhausted: %v", maxRetries, lastErr))
	return fmt.Errorf("nntp job failed after %d retries: %w", maxRetries, lastErr)
}

// Run starts the worker processing jobs from its job queue
func (w *Worker) Run(ctx context.Context, jobQueue <-chan *models.IndexingJob) {
	w.log.Info("Worker started", "worker_id", w.ID)

	for job := range jobQueue {
		// Check context cancellation before processing
		if ctx.Err() != nil {
			w.log.Info("Worker shutting down, discarding job",
				"worker_id", w.ID,
				"job_id", job.ID,
			)
			continue
		}

		w.log.Info("Worker processing job",
			"worker_id", w.ID,
			"job_id", job.ID,
			"job_type", job.JobType,
		)

		// Process the job with a timeout context
		jobCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		err := w.ProcessJob(jobCtx, job)
		cancel()

		if err != nil {
			w.log.Error("Worker failed processing job",
				"worker_id", w.ID,
				"job_id", job.ID,
				"error", err,
			)

			// Update job error in database
			if err := w.IndexingJobRepo.UpdateError(ctx, job.ID, err.Error()); err != nil {
				w.log.Error("Failed to update job error", "job_id", job.ID, "error", err)
			}
		} else {
			w.log.Info("Worker completed job",
				"worker_id", w.ID,
				"job_id", job.ID,
				"records_processed", job.RecordsProcessed,
				"records_failed", job.RecordsFailed,
			)
		}
	}

	w.log.Info("Worker stopped", "worker_id", w.ID)
}

// fetchPosts retrieves new posts from a Spotnet source
func (w *Worker) fetchPosts(ctx context.Context, source *models.SpotnetSource, sinceDate time.Time) ([]*spotnet_client.SpotnetPost, error) {
	if w.SpotnetClient == nil {
		return nil, fmt.Errorf("spotnet client not initialized")
	}

	posts, err := w.SpotnetClient.GetPosts(ctx, source, "", sinceDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	return posts, nil
}

// storePosts processes, parses, and bulk inserts posts
func (w *Worker) storePosts(ctx context.Context, job *models.IndexingJob, rawPosts []*spotnet_client.SpotnetPost) (int64, int64, error) {
	if len(rawPosts) == 0 {
		return 0, 0, nil
	}

	var processed int64
	var failed int64
	var dbPosts []*models.SpotnetPost

	for _, rawPost := range rawPosts {
		// Check for duplicate by external ID
		exists, err := w.PostRepo.ExistsByExternalID(ctx, rawPost.ID)
		if err != nil {
			w.log.Error("Failed to check post existence", "external_id", rawPost.ID, "error", err)
			failed++
			continue
		}
		if exists {
			// Skip duplicates
			continue
		}

		// Parse the raw post into a database model
		var dbPost *models.SpotnetPost

		if w.PostParser != nil {
			// Use the PostParser to convert raw data
			parsed, err := w.PostParser.ParseSpotnetPost(ctx, map[string]interface{}{
				"id":          rawPost.ID,
				"title":       rawPost.Title,
				"description": rawPost.Description,
				"author":      rawPost.Author,
				"category":    rawPost.Category,
				"posted_at":   rawPost.PostedAt.Format(time.RFC3339),
			}, nil)
			if err != nil {
				w.log.Error("Failed to parse post", "external_id", rawPost.ID, "error", err)
				failed++
				continue
			}
			dbPost = parsed
		} else {
			// Fallback: create directly from raw post data
			dbPost = &models.SpotnetPost{
				ExternalID:         rawPost.ID,
				Title:              rawPost.Title,
				Description:        rawPost.Description,
				Author:             rawPost.Author,
				CategoryName:       rawPost.Category,
				PostedAt:           rawPost.PostedAt,
				RawMetadata:        fmt.Sprintf(`{"id":"%s","category":"%s","url":"%s"}`, rawPost.ID, rawPost.Category, rawPost.URL),
				NormalizedMetadata: "{}",
			}
		}

		// Run metadata processing
		if w.MetadataProcessor != nil {
			if err := w.MetadataProcessor.ProcessPostMetadata(ctx, dbPost); err != nil {
				w.log.Error("Failed to process post metadata", "external_id", rawPost.ID, "error", err)
				failed++
				continue
			}
		}

		dbPosts = append(dbPosts, dbPost)
		processed++

		// Periodically update job progress
		if processed%100 == 0 {
			progress := float64(processed) / float64(len(rawPosts)) * 100.0
			if err := w.IndexingJobRepo.UpdateProgress(ctx, job.ID, progress, processed, failed); err != nil {
				w.log.Error("Failed to update job progress", "job_id", job.ID, "error", err)
			}
		}
	}

	// Bulk insert all posts
	if len(dbPosts) > 0 {
		if err := w.PostRepo.CreateBulk(ctx, dbPosts); err != nil {
			return processed, failed, fmt.Errorf("failed to bulk insert %d posts: %w", len(dbPosts), err)
		}
	}

	return processed, failed, nil
}

// getNNTPServers retrieves enabled NNTP servers from the database
func (w *Worker) getNNTPServers(ctx context.Context) ([]*models.NNTPServer, error) {
	// Query enabled NNTP servers, ordered by priority descending
	// This would use an NNTPServerRepository in production
	// For now, return nil as a placeholder since there is no direct server repo
	return nil, nil
}

// fetchArticles retrieves new articles from a newsgroup via NNTP
func (w *Worker) fetchArticles(ctx context.Context, newsgroup *models.Newsgroup, sinceArticleNum int64) ([]*models.Article, error) {
	// In a real implementation, this would:
	// 1. Get an NNTP connection from the ConnectionManager
	// 2. Use NNTP commands (LIST, STAT, BODY, etc.) to fetch articles
	// 3. Parse the article headers and body
	// 4. Return a slice of Article models
	//
	// Example implementation:
	//
	// servers, err := w.getNNTPServers(ctx)
	// if err != nil {
	//     return nil, err
	// }
	//
	// server := servers[0] // Use highest priority server
	// conn, err := w.ConnectionManager.GetConnection(ctx, server)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get NNTP connection: %w", err)
	// }
	// defer w.ConnectionManager.ReleaseConnection(server.ID, conn)
	//
	// // Select the newsgroup
	// conn.SendCommand(ctx, "GROUP", newsgroup.Name)
	//
	// // Fetch articles since last indexed
	// start := sinceArticleNum + 1
	// end := newsgroup.LastArticleNumber
	// if start > end {
	//     return nil, nil
	// }
	//
	// // Use XOVER to get article metadata efficiently
	// var articles []*models.Article
	// for num := start; num <= end; num++ {
	//     article, err := conn.FetchArticle(ctx, newsgroup.Name, num)
	//     if err != nil {
	//         // Some articles may not exist; continue
	//         continue
	//     }
	//     articles = append(articles, article)
	// }
	//
	// return articles, nil

	return nil, nil
}

// storeArticles processes and bulk inserts articles
func (w *Worker) storeArticles(ctx context.Context, job *models.IndexingJob, newsgroup *models.Newsgroup, articles []*models.Article) (int64, int64, error) {
	if len(articles) == 0 {
		return 0, 0, nil
	}

	var processed int64
	var failed int64

	for _, article := range articles {
		// Check for duplicate by message ID
		exists, err := w.ArticleRepo.ExistsByMessageID(ctx, article.MessageID)
		if err != nil {
			w.log.Error("Failed to check article existence", "message_id", article.MessageID, "error", err)
			failed++
			continue
		}
		if exists {
			processed++
			continue
		}

		processed++

		// Periodically update job progress
		if processed%100 == 0 {
			progress := float64(processed) / float64(len(articles)) * 100.0
			if err := w.IndexingJobRepo.UpdateProgress(ctx, job.ID, progress, processed, failed); err != nil {
				w.log.Error("Failed to update job progress", "job_id", job.ID, "error", err)
			}
		}
	}

	// Bulk insert all articles
	if err := w.ArticleRepo.CreateBulk(ctx, articles); err != nil {
		return processed, failed, fmt.Errorf("failed to bulk insert %d articles: %w", len(articles), err)
	}

	return processed, failed, nil
}

// markJobFailed updates the job status to failed with an error message
func (w *Worker) markJobFailed(job *models.IndexingJob, errorMessage string) {
	now := time.Now().UTC()
	job.Status = "failed"
	job.ErrorMessage = errorMessage
	job.CompletedAt = &now

	if err := w.IndexingJobRepo.UpdateError(context.Background(), job.ID, errorMessage); err != nil {
		w.log.Error("Failed to update job error status", "job_id", job.ID, "error", err)
	}
}
