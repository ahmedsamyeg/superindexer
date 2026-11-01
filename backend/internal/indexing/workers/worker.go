package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/spotnet/client"
	"github.com/superindexer/indexer/internal/spotnet/metadata"
	"github.com/superindexer/indexer/internal/spotnet/parser"
	"github.com/superindexer/indexer/internal/usenet/client"
)

// Worker represents an indexing worker that processes jobs
type Worker struct {
	ID                int64
	SourceRepo        repositories.SpotnetSourceRepository
	PostRepo          repositories.SpotnetPostRepository
	ArticleRepo       repositories.ArticleRepository
	IndexingJobRepo   repositories.IndexingJobRepository
	SpotnetClient     *client.SpotnetClient
	UsenetClient      *usenet_client.NNTPClient // This is a placeholder - need to verify actual package name
	PostParser        *parser.PostParser
	MetadataProcessor *metadata.MetadataProcessor
}

// NewWorker creates a new indexing worker
func NewWorker(
	id int64,
	sourceRepo repositories.SpotnetSourceRepository,
	postRepo repositories.SpotnetPostRepository,
	articleRepo repositories.ArticleRepository,
	indexingJobRepo repositories.IndexingJobRepository,
	spotnetClient *client.SpotnetClient,
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
		SpotnetClient:     spotnetClient,
		UsenetClient:      usenetClient,
		PostParser:        postParser,
		MetadataProcessor: metadataProcessor,
	}
}

// ProcessJob processes an indexing job
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

// ProcessSpotnetJob processes a Spotnet indexing job
func (w *Worker) ProcessSpotnetJob(ctx context.Context, job *models.IndexingJob) error {
	// Get the source
	source, err := w.SourceRepo.GetByID(ctx, job.SourceID)
	if err != nil {
		return fmt.Errorf("failed to get source: %w", err)
	}

	// Update job status
	job.Status = "running"
	job.StartedAt = &time.Now()
	err = w.IndexingJobRepo.Update(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Synchronize source
	processed, err := w.SpotnetClient.SyncSource(ctx, source)
	if err != nil {
		job.ErrorMessage = err.Error()
		job.Status = "failed"
		err = w.IndexingJobRepo.Update(ctx, job)
		if err != nil {
			return fmt.Errorf("failed to update failed job: %w", err)
		}
		return err
	}

	// Update job completion
	job.Status = "completed"
	job.RecordsProcessed = processed
	job.CompletedAt = &time.Now()
	err = w.IndexingJobRepo.Update(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update completed job: %w", err)
	}

	return nil
}

// ProcessUsenetJob processes a Usenet indexing job (placeholder)
func (w *Worker) ProcessUsenetJob(ctx context.Context, job *models.IndexingJob) error {
	// This would implement actual NNTP article processing logic
	return nil
}