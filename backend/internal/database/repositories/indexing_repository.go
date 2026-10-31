package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// IndexingJobRepository defines the interface for indexing job data access.
type IndexingJobRepository interface {
	Create(ctx context.Context, job *models.IndexingJob) error
	GetByID(ctx context.Context, id int64) (*models.IndexingJob, error)
	GetByStatus(ctx context.Context, status string) ([]*models.IndexingJob, error)
	GetPendingJobs(ctx context.Context) ([]*models.IndexingJob, error)
	Update(ctx context.Context, job *models.IndexingJob) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateProgress(ctx context.Context, id int64, progress float64, processed int64, failed int64) error
	UpdateError(ctx context.Context, id int64, errorMessage string) error
	Complete(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountActive(ctx context.Context) (int64, error)
}

// indexingJobRepository implements IndexingJobRepository.
type indexingJobRepository struct {
	db *bun.DB
}

// NewIndexingJobRepository creates a new IndexingJobRepository.
func NewIndexingJobRepository(db *bun.DB) IndexingJobRepository {
	return &indexingJobRepository{db: db}
}

func (r *indexingJobRepository) Create(ctx context.Context, job *models.IndexingJob) error {
	_, err := r.db.NewInsert().Model(job).Exec(ctx)
	return err
}

func (r *indexingJobRepository) GetByID(ctx context.Context, id int64) (*models.IndexingJob, error) {
	job := &models.IndexingJob{}
	err := r.db.NewSelect().Model(job).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (r *indexingJobRepository) GetByStatus(ctx context.Context, status string) ([]*models.IndexingJob, error) {
	var jobs []*models.IndexingJob
	err := r.db.NewSelect().Model(&jobs).
		Where("status = ?", status).
		Order("priority DESC, created_at ASC").
		Scan(ctx)
	return jobs, err
}

func (r *indexingJobRepository) GetPendingJobs(ctx context.Context) ([]*models.IndexingJob, error) {
	return r.GetByStatus(ctx, "pending")
}

func (r *indexingJobRepository) Update(ctx context.Context, job *models.IndexingJob) error {
	_, err := r.db.NewUpdate().Model(job).Where("id = ?", job.ID).Exec(ctx)
	return err
}

func (r *indexingJobRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := r.db.NewUpdate().
		Model(&models.IndexingJob{}).
		Set("status = ?", status).
		Where("id = ?", id)

	if status == "running" {
		query = query.Set("started_at = ?", time.Now().UTC())
	} else if status == "completed" || status == "failed" {
		query = query.Set("completed_at = ?", time.Now().UTC())
	}

	_, err := query.Exec(ctx)
	return err
}

func (r *indexingJobRepository) UpdateProgress(ctx context.Context, id int64, progress float64, processed int64, failed int64) error {
	_, err := r.db.NewUpdate().
		Model(&models.IndexingJob{}).
		Set("progress = ?", progress).
		Set("records_processed = ?", processed).
		Set("records_failed = ?", failed).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *indexingJobRepository) UpdateError(ctx context.Context, id int64, errorMessage string) error {
	_, err := r.db.NewUpdate().
		Model(&models.IndexingJob{}).
		Set("error_message = ?", errorMessage).
		Set("status = ?", "failed").
		Set("completed_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *indexingJobRepository) Complete(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		Model(&models.IndexingJob{}).
		Set("status = ?", "completed").
		Set("progress = ?", 100.0).
		Set("completed_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *indexingJobRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().Model(&models.IndexingJob{}).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("indexing job not found: %d", id)
	}
	return nil
}

func (r *indexingJobRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	return r.db.NewSelect().Model(&models.IndexingJob{}).
		Where("status = ?", status).
		Count(ctx)
}

func (r *indexingJobRepository) CountActive(ctx context.Context) (int64, error) {
	return r.db.NewSelect().Model(&models.IndexingJob{}).
		Where("status IN (?, ?)", "pending", "running").
		Count(ctx)
}

// Ensure interface compliance
var _ IndexingJobRepository = (*indexingJobRepository)(nil)

// CheckpointRepository defines the interface for checkpoint data access.
type CheckpointRepository interface {
	CreateOrUpdate(ctx context.Context, checkpoint *models.IndexingCheckpoint) error
	Get(ctx context.Context, sourceType string, sourceID int64, newsgroupID *int64) (*models.IndexingCheckpoint, error)
	Delete(ctx context.Context, sourceType string, sourceID int64, newsgroupID *int64) error
}

// checkpointRepository implements CheckpointRepository.
type checkpointRepository struct {
	db *bun.DB
}

// NewCheckpointRepository creates a new CheckpointRepository.
func NewCheckpointRepository(db *bun.DB) CheckpointRepository {
	return &checkpointRepository{db: db}
}

func (r *checkpointRepository) CreateOrUpdate(ctx context.Context, checkpoint *models.IndexingCheckpoint) error {
	_, err := r.db.NewInsert().
		Model(checkpoint).
		On("CONFLICT (source_type, source_id, newsgroup_id) DO UPDATE").
		Set("last_processed_position = EXCLUDED.last_processed_position").
		Set("last_processed_at = EXCLUDED.last_processed_at").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)
	return err
}

func (r *checkpointRepository) Get(ctx context.Context, sourceType string, sourceID int64, newsgroupID *int64) (*models.IndexingCheckpoint, error) {
	checkpoint := &models.IndexingCheckpoint{}
	query := r.db.NewSelect().Model(checkpoint).
		Where("source_type = ?", sourceType).
		Where("source_id = ?", sourceID)

	if newsgroupID != nil {
		query = query.Where("newsgroup_id = ?", *newsgroupID)
	} else {
		query = query.Where("newsgroup_id IS NULL")
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return checkpoint, nil
}

func (r *checkpointRepository) Delete(ctx context.Context, sourceType string, sourceID int64, newsgroupID *int64) error {
	query := r.db.NewDelete().
		Model(&models.IndexingCheckpoint{}).
		Where("source_type = ?", sourceType).
		Where("source_id = ?", sourceID)

	if newsgroupID != nil {
		query = query.Where("newsgroup_id = ?", *newsgroupID)
	} else {
		query = query.Where("newsgroup_id IS NULL")
	}

	_, err := query.Exec(ctx)
	return err
}

// Ensure interface compliance
var _ CheckpointRepository = (*checkpointRepository)(nil)
