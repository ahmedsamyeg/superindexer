package repositories

import (
	"context"
	"fmt"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// SpotnetSourceRepository defines the interface for Spotnet source data access.
type SpotnetSourceRepository interface {
	Create(ctx context.Context, source *models.SpotnetSource) error
	GetByID(ctx context.Context, id int64) (*models.SpotnetSource, error)
	GetAll(ctx context.Context) ([]*models.SpotnetSource, error)
	GetEnabled(ctx context.Context) ([]*models.SpotnetSource, error)
	Update(ctx context.Context, source *models.SpotnetSource) error
	Delete(ctx context.Context, id int64) error
	UpdateLastSync(ctx context.Context, id int64, lastSyncAt interface{}) error
}

// spotnetSourceRepository implements SpotnetSourceRepository.
type spotnetSourceRepository struct {
	db *bun.DB
}

// NewSpotnetSourceRepository creates a new SpotnetSourceRepository.
func NewSpotnetSourceRepository(db *bun.DB) SpotnetSourceRepository {
	return &spotnetSourceRepository{db: db}
}

func (r *spotnetSourceRepository) Create(ctx context.Context, source *models.SpotnetSource) error {
	_, err := r.db.NewInsert().Model(source).Exec(ctx)
	return err
}

func (r *spotnetSourceRepository) GetByID(ctx context.Context, id int64) (*models.SpotnetSource, error) {
	source := &models.SpotnetSource{}
	err := r.db.NewSelect().Model(source).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return source, nil
}

func (r *spotnetSourceRepository) GetAll(ctx context.Context) ([]*models.SpotnetSource, error) {
	var sources []*models.SpotnetSource
	err := r.db.NewSelect().Model(&sources).Order("name ASC").Scan(ctx)
	return sources, err
}

func (r *spotnetSourceRepository) GetEnabled(ctx context.Context) ([]*models.SpotnetSource, error) {
	var sources []*models.SpotnetSource
	err := r.db.NewSelect().Model(&sources).
		Where("enabled = true").
		Order("last_sync_at DESC").
		Scan(ctx)
	return sources, err
}

func (r *spotnetSourceRepository) Update(ctx context.Context, source *models.SpotnetSource) error {
	_, err := r.db.NewUpdate().Model(source).Where("id = ?", source.ID).Exec(ctx)
	return err
}

func (r *spotnetSourceRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().Model(&models.SpotnetSource{}).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("spotnet source not found: %d", id)
	}
	return nil
}

func (r *spotnetSourceRepository) UpdateLastSync(ctx context.Context, id int64, lastSyncAt interface{}) error {
	_, err := r.db.NewUpdate().
		Model(&models.SpotnetSource{}).
		Set("last_sync_at = ?", lastSyncAt).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// Ensure interface compliance
var _ SpotnetSourceRepository = (*spotnetSourceRepository)(nil)

// SpotnetPostRepository defines the interface for Spotnet post data access.
type SpotnetPostRepository interface {
	Create(ctx context.Context, post *models.SpotnetPost) error
	CreateBulk(ctx context.Context, posts []*models.SpotnetPost) error
	GetByID(ctx context.Context, id int64) (*models.SpotnetPost, error)
	GetByExternalID(ctx context.Context, externalID string) (*models.SpotnetPost, error)
	ExistsByExternalID(ctx context.Context, externalID string) (bool, error)
	Update(ctx context.Context, post *models.SpotnetPost) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	CountBySource(ctx context.Context, sourceID int64) (int64, error)
	ListBySource(ctx context.Context, sourceID int64, pagination Pagination) (*PaginatedResult, error)
}

// spotnetPostRepository implements SpotnetPostRepository.
type spotnetPostRepository struct {
	db *bun.DB
}

// NewSpotnetPostRepository creates a new SpotnetPostRepository.
func NewSpotnetPostRepository(db *bun.DB) SpotnetPostRepository {
	return &spotnetPostRepository{db: db}
}

func (r *spotnetPostRepository) Create(ctx context.Context, post *models.SpotnetPost) error {
	_, err := r.db.NewInsert().Model(post).Exec(ctx)
	return err
}

func (r *spotnetPostRepository) CreateBulk(ctx context.Context, posts []*models.SpotnetPost) error {
	if len(posts) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().
		Model(&posts).
		On("CONFLICT (external_id) DO NOTHING").
		Exec(ctx)
	return err
}

func (r *spotnetPostRepository) GetByID(ctx context.Context, id int64) (*models.SpotnetPost, error) {
	post := &models.SpotnetPost{}
	err := r.db.NewSelect().Model(post).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *spotnetPostRepository) GetByExternalID(ctx context.Context, externalID string) (*models.SpotnetPost, error) {
	post := &models.SpotnetPost{}
	err := r.db.NewSelect().Model(post).Where("external_id = ?", externalID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *spotnetPostRepository) ExistsByExternalID(ctx context.Context, externalID string) (bool, error) {
	var count int64
	err := r.db.NewSelect().
		Model(&models.SpotnetPost{}).
		Column("count(*)").
		Where("external_id = ?", externalID).
		Scan(ctx, &count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *spotnetPostRepository) Update(ctx context.Context, post *models.SpotnetPost) error {
	_, err := r.db.NewUpdate().Model(post).Where("id = ?", post.ID).Exec(ctx)
	return err
}

func (r *spotnetPostRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().Model(&models.SpotnetPost{}).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("spotnet post not found: %d", id)
	}
	return nil
}

func (r *spotnetPostRepository) Count(ctx context.Context) (int64, error) {
	return r.db.NewSelect().Model(&models.SpotnetPost{}).Count(ctx)
}

func (r *spotnetPostRepository) CountBySource(ctx context.Context, sourceID int64) (int64, error) {
	return r.db.NewSelect().Model(&models.SpotnetPost{}).
		Where("source_id = ?", sourceID).
		Count(ctx)
}

func (r *spotnetPostRepository) ListBySource(ctx context.Context, sourceID int64, pagination Pagination) (*PaginatedResult, error) {
	var posts []models.SpotnetPost
	query := r.db.NewSelect().Model(&posts).
		Where("source_id = ?", sourceID)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	query = query.
		Offset(pagination.Offset()).
		Limit(pagination.PerPage).
		Order("posted_at DESC")

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return newUserPaginatedResult(posts, total, int64(pagination.Page), int64(pagination.PerPage)), nil
}

// Ensure interface compliance
var _ SpotnetPostRepository = (*spotnetPostRepository)(nil)
