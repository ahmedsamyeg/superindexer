package repositories

import (
	"context"
	"fmt"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// CategoryRepository defines the interface for category data access.
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id int64) (*models.Category, error)
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	GetByName(ctx context.Context, name string) (*models.Category, error)
	GetAll(ctx context.Context) ([]*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id int64) error
	GetChildren(ctx context.Context, parentID int64) ([]*models.Category, error)
}

// categoryRepository implements CategoryRepository.
type categoryRepository struct {
	db *bun.DB
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *bun.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	_, err := r.db.NewInsert().Model(category).Exec(ctx)
	return err
}

func (r *categoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	category := &models.Category{}
	err := r.db.NewSelect().Model(category).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	category := &models.Category{}
	err := r.db.NewSelect().Model(category).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) GetByName(ctx context.Context, name string) (*models.Category, error) {
	category := &models.Category{}
	err := r.db.NewSelect().Model(category).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.NewSelect().Model(&categories).Order("name ASC").Scan(ctx)
	return categories, err
}

func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	_, err := r.db.NewUpdate().Model(category).Where("id = ?", category.ID).Exec(ctx)
	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().Model(&models.Category{}).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("category not found: %d", id)
	}
	return nil
}

func (r *categoryRepository) GetChildren(ctx context.Context, parentID int64) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.NewSelect().Model(&categories).
		Where("parent_id = ?", parentID).
		Order("name ASC").
		Scan(ctx)
	return categories, err
}

// Ensure interface compliance
var _ CategoryRepository = (*categoryRepository)(nil)
