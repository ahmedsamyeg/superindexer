package repositories

import (
	"context"
	"fmt"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// ArticleRepository defines the interface for article data access.
type ArticleRepository interface {
	Create(ctx context.Context, article *models.Article) error
	CreateBulk(ctx context.Context, articles []*models.Article) error
	GetByID(ctx context.Context, id int64) (*models.Article, error)
	GetByMessageID(ctx context.Context, messageID string) (*models.Article, error)
	Update(ctx context.Context, article *models.Article) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	CountByNewsgroup(ctx context.Context, newsgroupID int64) (int64, error)
	ListByNewsgroup(ctx context.Context, newsgroupID int64, pagination Pagination) (*PaginatedResult, error)
	ExistsByMessageID(ctx context.Context, messageID string) (bool, error)
}

// articleRepository implements ArticleRepository using Bun.
type articleRepository struct {
	db *bun.DB
}

// NewArticleRepository creates a new ArticleRepository.
func NewArticleRepository(db *bun.DB) ArticleRepository {
	return &articleRepository{db: db}
}

// Create inserts a new article.
func (r *articleRepository) Create(ctx context.Context, article *models.Article) error {
	_, err := r.db.NewInsert().Model(article).Exec(ctx)
	return err
}

// CreateBulk inserts multiple articles efficiently using ON CONFLICT DO NOTHING.
func (r *articleRepository) CreateBulk(ctx context.Context, articles []*models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	_, err := r.db.NewInsert().
		Model(&articles).
		On("CONFLICT (message_id) DO NOTHING").
		Exec(ctx)
	return err
}

// GetByID retrieves an article by ID.
func (r *articleRepository) GetByID(ctx context.Context, id int64) (*models.Article, error) {
	article := &models.Article{}
	err := r.db.NewSelect().Model(article).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return article, nil
}

// GetByMessageID retrieves an article by message ID.
func (r *articleRepository) GetByMessageID(ctx context.Context, messageID string) (*models.Article, error) {
	article := &models.Article{}
	err := r.db.NewSelect().Model(article).
		Where("message_id = ?", messageID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return article, nil
}

// Update updates an existing article.
func (r *articleRepository) Update(ctx context.Context, article *models.Article) error {
	_, err := r.db.NewUpdate().Model(article).Where("id = ?", article.ID).Exec(ctx)
	return err
}

// Delete removes an article by ID.
func (r *articleRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().
		Model(&models.Article{}).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("article not found: %d", id)
	}
	return nil
}

// Count returns the total number of articles.
func (r *articleRepository) Count(ctx context.Context) (int64, error) {
	return r.db.NewSelect().Model(&models.Article{}).Count(ctx)
}

// CountByNewsgroup returns the number of articles in a newsgroup.
func (r *articleRepository) CountByNewsgroup(ctx context.Context, newsgroupID int64) (int64, error) {
	return r.db.NewSelect().Model(&models.Article{}).
		Where("newsgroup_id = ?", newsgroupID).
		Count(ctx)
}

// ListByNewsgroup retrieves paginated articles for a newsgroup.
func (r *articleRepository) ListByNewsgroup(ctx context.Context, newsgroupID int64, pagination Pagination) (*PaginatedResult, error) {
	var articles []models.Article
	query := r.db.NewSelect().Model(&articles).
		Where("newsgroup_id = ?", newsgroupID)

	// Count total
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	query = query.
		Offset(int64((pagination.Page - 1) * pagination.PerPage)).
		Limit(int64(pagination.PerPage)).
		Order("posted_at DESC")

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return newUserPaginatedResult(articles, total, int64(pagination.Page), int64(pagination.PerPage)), nil
}

// ExistsByMessageID checks if an article with the given message ID exists.
func (r *articleRepository) ExistsByMessageID(ctx context.Context, messageID string) (bool, error) {
	var count int64
	err := r.db.NewSelect().Model(&models.Article{}).
		Column("count(*)").
		Where("message_id = ?", messageID).
		Scan(ctx, &count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Ensure interface compliance
var _ ArticleRepository = (*articleRepository)(nil)
