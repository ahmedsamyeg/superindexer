package repositories

import (
	"context"
	"fmt"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// NewsgroupRepository defines the interface for newsgroup data access.
type NewsgroupRepository interface {
	Create(ctx context.Context, group *models.Newsgroup) error
	GetByID(ctx context.Context, id int64) (*models.Newsgroup, error)
	GetByName(ctx context.Context, name string) (*models.Newsgroup, error)
	GetAll(ctx context.Context) ([]*models.Newsgroup, error)
	GetEnabled(ctx context.Context) ([]*models.Newsgroup, error)
	Update(ctx context.Context, group *models.Newsgroup) error
	UpdateArticleRange(ctx context.Context, id int64, first, last int64) error
	UpdateLastIndexed(ctx context.Context, id int64, articleNum int64, timestamp interface{}) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	CountEnabled(ctx context.Context) (int64, error)
}

// newsgroupRepository implements NewsgroupRepository.
type newsgroupRepository struct {
	db *bun.DB
}

// NewNewsgroupRepository creates a new NewsgroupRepository.
func NewNewsgroupRepository(db *bun.DB) NewsgroupRepository {
	return &newsgroupRepository{db: db}
}

func (r *newsgroupRepository) Create(ctx context.Context, group *models.Newsgroup) error {
	_, err := r.db.NewInsert().Model(group).Exec(ctx)
	return err
}

func (r *newsgroupRepository) GetByID(ctx context.Context, id int64) (*models.Newsgroup, error) {
	group := &models.Newsgroup{}
	err := r.db.NewSelect().Model(group).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *newsgroupRepository) GetByName(ctx context.Context, name string) (*models.Newsgroup, error) {
	group := &models.Newsgroup{}
	err := r.db.NewSelect().Model(group).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *newsgroupRepository) GetAll(ctx context.Context) ([]*models.Newsgroup, error) {
	var groups []*models.Newsgroup
	err := r.db.NewSelect().Model(&groups).Order("name ASC").Scan(ctx)
	return groups, err
}

func (r *newsgroupRepository) GetEnabled(ctx context.Context) ([]*models.Newsgroup, error) {
	var groups []*models.Newsgroup
	err := r.db.NewSelect().Model(&groups).
		Where("enabled = true").
		Order("indexing_priority DESC, name ASC").
		Scan(ctx)
	return groups, err
}

func (r *newsgroupRepository) Update(ctx context.Context, group *models.Newsgroup) error {
	_, err := r.db.NewUpdate().Model(group).Where("id = ?", group.ID).Exec(ctx)
	return err
}

func (r *newsgroupRepository) UpdateArticleRange(ctx context.Context, id int64, first, last int64) error {
	_, err := r.db.NewUpdate().
		Model(&models.Newsgroup{}).
		Set("first_article_number = ?", first).
		Set("last_article_number = ?", last).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *newsgroupRepository) UpdateLastIndexed(ctx context.Context, id int64, articleNum int64, timestamp interface{}) error {
	query := r.db.NewUpdate().
		Model(&models.Newsgroup{}).
		Set("last_indexed_article_number = ?", articleNum).
		Where("id = ?", id)

	if timestamp != nil {
		query = query.Set("last_indexed_at = ?", timestamp)
	}

	_, err := query.Exec(ctx)
	return err
}

func (r *newsgroupRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().Model(&models.Newsgroup{}).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("newsgroup not found: %d", id)
	}
	return nil
}

func (r *newsgroupRepository) Count(ctx context.Context) (int64, error) {
	return r.db.NewSelect().Model(&models.Newsgroup{}).Count(ctx)
}

func (r *newsgroupRepository) CountEnabled(ctx context.Context) (int64, error) {
	return r.db.NewSelect().Model(&models.Newsgroup{}).Where("enabled = true").Count(ctx)
}

// Ensure interface compliance
var _ NewsgroupRepository = (*newsgroupRepository)(nil)
