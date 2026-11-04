package service

import (
	"context"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/indexing/coordinator"
	"github.com/superindexer/indexer/internal/indexing/workers"
	"log/slog"
)

// IndexingService handles the indexing engine operations
type IndexingService struct {
	jobRepo       repositories.IndexingJobRepository
	sourceRepo    repositories.SpotnetSourceRepository
	newsgroupRepo repositories.NewsgroupRepository
	articleRepo   repositories.ArticleRepository
	postRepo      repositories.SpotnetPostRepository
	coordinator   *coordinator.IndexingCoordinator
	workerPool    *workers.WorkerPool
	log           *slog.Logger
}

// NewIndexingService creates a new indexing service
func NewIndexingService(
	jobRepo repositories.IndexingJobRepository,
	sourceRepo repositories.SpotnetSourceRepository,
	newsgroupRepo repositories.NewsgroupRepository,
	articleRepo repositories.ArticleRepository,
	postRepo repositories.SpotnetPostRepository,
	coordinator *coordinator.IndexingCoordinator,
	workerPool *workers.WorkerPool,
) *IndexingService {
	return &IndexingService{
		jobRepo:       jobRepo,
		sourceRepo:    sourceRepo,
		newsgroupRepo: newsgroupRepo,
		articleRepo:   articleRepo,
		postRepo:      postRepo,
		coordinator:   coordinator,
		workerPool:    workerPool,
		log:           slog.Default(),
	}
}

// CreateIndexingJob creates and schedules a new indexing job
func (s *IndexingService) CreateIndexingJob(ctx context.Context, jobType string, sourceID, newsgroupID *int64) (*models.IndexingJob, error) {
	// Validate job type
	if jobType != "spotnet" && jobType != "nntp" {
		return nil, fmt.Errorf("invalid job type: %s", jobType)
	}

	// Validate source/newsgroup requirements
	switch jobType {
	case "spotnet":
		if sourceID == nil {
			return nil, fmt.Errorf("spotnet jobs require a source ID")
		}
		// Verify source exists and is enabled
		source, err := s.sourceRepo.GetByID(ctx, *sourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get source %d: %w", *sourceID, err)
		}
		if !source.Enabled {
			return nil, fmt.Errorf("source %d is disabled", *sourceID)
		}
	case "nntp":
		if newsgroupID == nil {
			return nil, fmt.Errorf("nntp jobs require a newsgroup ID")
		}
		// Verify newsgroup exists and is enabled
		newsgroup, err := s.newsgroupRepo.GetByID(ctx, *newsgroupID)
		if err != nil {
			return nil, fmt.Errorf("failed to get newsgroup %d: %w", *newsgroupID, err)
		}
		if !newsgroup.Enabled {
			return nil, fmt.Errorf("newsgroup %d is disabled", *newsgroupID)
		}
	}

	// Submit job to coordinator
	job, err := s.coordinator.SubmitJob(ctx, jobType, sourceID, newsgroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to submit indexing job: %w", err)
	}

	s.log.Info("Indexing job created",
		"job_id", job.ID,
		"job_type", jobType,
		"source_id", sourceID,
		"newsgroup_id", newsgroupID,
	)

	return job, nil
}

// StartIndexingEngine initializes and starts the indexing engine
func (s *IndexingService) StartIndexingEngine(ctx context.Context) error {
	s.log.Info("Starting indexing engine")

	// Start the coordinator (which also starts the worker pool)
	if err := s.coordinator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start indexing coordinator: %w", err)
	}

	s.log.Info("Indexing engine started successfully")
	return nil
}

// StopIndexingEngine shuts down the indexing engine gracefully
func (s *IndexingService) StopIndexingEngine(ctx context.Context) error {
	s.log.Info("Stopping indexing engine")

	// Stop the coordinator (which also stops the worker pool)
	if err := s.coordinator.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop indexing coordinator: %w", err)
	}

	s.log.Info("Indexing engine stopped")
	return nil
}

// ListIndexingJobs retrieves a paginated list of indexing jobs
func (s *IndexingService) ListIndexingJobs(ctx context.Context, status string, page, perPage int) (*repositories.PaginatedResult, error) {
	// Delegate to coordinator for job listing
	return s.coordinator.ListJobs(ctx, status, page, perPage)
}

// GetIndexingJob retrieves details of a specific indexing job
func (s *IndexingService) GetIndexingJob(ctx context.Context, jobID int64) (*models.IndexingJob, error) {
	return s.coordinator.GetJobStatus(ctx, jobID)
}

// CancelIndexingJob cancels a pending or running indexing job
func (s *IndexingService) CancelIndexingJob(ctx context.Context, jobID int64) error {
	// Get current job status
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job %d: %w", jobID, err)
	}

	// Check if job can be cancelled
	if job.Status == "completed" || job.Status == "cancelled" {
		return fmt.Errorf("cannot cancel job in %s status", job.Status)
	}
	if job.Status == "failed" {
		return fmt.Errorf("cannot cancel a failed job; use RetryIndexingJob instead")
	}

	// Cancel via coordinator
	if err := s.coordinator.CancelJob(ctx, jobID); err != nil {
		return fmt.Errorf("failed to cancel job %d: %w", jobID, err)
	}

	s.log.Info("Indexing job cancelled", "job_id", jobID)
	return nil
}

// RetryIndexingJob retries a failed indexing job
func (s *IndexingService) RetryIndexingJob(ctx context.Context, jobID int64) error {
	// Get current job status
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job %d: %w", jobID, err)
	}

	// Only retry failed jobs
	if job.Status != "failed" {
		return fmt.Errorf("can only retry failed jobs; current status: %s", job.Status)
	}

	// Retry via coordinator
	if err := s.coordinator.RetryJob(ctx, jobID); err != nil {
		return fmt.Errorf("failed to retry job %d: %w", jobID, err)
	}

	s.log.Info("Indexing job retried", "job_id", jobID)
	return nil
}

// GetIndexingStats retrieves statistics about indexing operations
func (s *IndexingService) GetIndexingStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count jobs by status
	statuses := []string{"pending", "running", "completed", "failed", "cancelled"}
	for _, status := range statuses {
		count, err := s.jobRepo.CountByStatus(ctx, status)
		if err != nil {
			s.log.Error("Failed to count jobs by status", "status", status, "error", err)
			continue
		}
		stats[fmt.Sprintf("%s_jobs", status)] = count
	}

	// Count active jobs (pending + running)
	activeCount, err := s.jobRepo.CountActive(ctx)
	if err != nil {
		s.log.Error("Failed to count active jobs", "error", err)
	} else {
		stats["active_jobs"] = activeCount
	}

	// Count total jobs
	totalCount, err := s.jobRepo.CountByStatus(ctx, "")
	if err != nil {
		// Fallback: count all statuses
		var total int64
		for _, status := range statuses {
			count, e := s.jobRepo.CountByStatus(ctx, status)
			if e != nil {
				continue
			}
			total += count
		}
		stats["total_jobs"] = total
	} else {
		stats["total_jobs"] = totalCount
	}

	// Count sources
	sourceCount, err := s.sourceRepo.GetEnabled(ctx)
	if err != nil {
		s.log.Error("Failed to count enabled sources", "error", err)
	} else {
		stats["enabled_sources"] = int64(len(sourceCount))
	}

	// Count newsgroups
	newsgroupCount, err := s.newsgroupRepo.CountEnabled(ctx)
	if err != nil {
		s.log.Error("Failed to count enabled newsgroups", "error", err)
	} else {
		stats["enabled_newsgroups"] = newsgroupCount
	}

	// Count articles and posts
	articleCount, err := s.articleRepo.Count(ctx)
	if err != nil {
		s.log.Error("Failed to count articles", "error", err)
	} else {
		stats["total_articles"] = articleCount
	}

	postCount, err := s.postRepo.Count(ctx)
	if err != nil {
		s.log.Error("Failed to count posts", "error", err)
	} else {
		stats["total_posts"] = postCount
	}

	return stats, nil
}

// ConfigureWorkerPool configures the worker pool for indexing jobs
func (s *IndexingService) ConfigureWorkerPool(ctx context.Context, maxWorkers int) error {
	if maxWorkers < 1 {
		return fmt.Errorf("maxWorkers must be at least 1")
	}

	// Stop current pool if running
	if s.workerPool != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Wait for in-flight jobs to finish
		if err := s.workerPool.Stop(shutdownCtx); err != nil {
			s.log.Error("Failed to stop existing worker pool for reconfiguration", "error", err)
		}
	}

	// Create new worker pool with the specified worker count
	newPool := workers.NewWorkerPool(maxWorkers)

	// Create workers for the pool
	workersList := make([]*workers.Worker, maxWorkers)
	for i := int64(0); i < int64(maxWorkers); i++ {
		workersList[i] = workers.NewWorker(
			i+1,
			s.sourceRepo,
			s.postRepo,
			s.articleRepo,
			s.jobRepo,
			s.newsgroupRepo,
			nil, // SpotnetClient - would be injected in production
			nil, // UsenetClient - would be injected in production
			nil, // PostParser - would be injected in production
			nil, // MetadataProcessor - would be injected in production
		)
	}

	// Update references
	s.workerPool = newPool

	s.log.Info("Worker pool reconfigured", "max_workers", maxWorkers)
	return nil
}
