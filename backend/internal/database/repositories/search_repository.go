package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// SearchRepository defines the interface for search operations.
type SearchRepository interface {
	SearchArticles(ctx context.Context, query models.SearchQuery) (*PaginatedResult, error)
	SearchSpotnetPosts(ctx context.Context, query models.SearchQuery) (*PaginatedResult, error)
	SearchAll(ctx context.Context, query models.SearchQuery) (*PaginatedResult, error)
	SearchArticlesByCategory(ctx context.Context, category string, pagination Pagination) (*PaginatedResult, error)
	GetFacets(ctx context.Context, query models.SearchQuery) (map[string]int64, error)
}

// searchRepository implements SearchRepository using PostgreSQL FTS.
type searchRepository struct {
	db *bun.DB
}

// NewSearchRepository creates a new SearchRepository.
func NewSearchRepository(db *bun.DB) SearchRepository {
	return &searchRepository{db: db}
}

// SearchArticles searches articles using PostgreSQL full-text search.
func (r *searchRepository) SearchArticles(ctx context.Context, query models.SearchQuery) (*PaginatedResult, error) {
	pagination := models.NewPagination(query.Page, query.PerPage)

	// Build the base query
	baseSQL := `
		SELECT a.id, a.message_id, a.newsgroup_id, n.name as newsgroup_name,
		       a.subject, a.author, a.posted_at, a.article_size,
		       a.normalized_metadata, a.created_at,
		       ts_rank_cd(a.search_vector, query) AS rank
		FROM articles a
		JOIN newsgroups n ON a.newsgroup_id = n.id,
		     websearch_to_tsquery('simple', $1) AS query
		WHERE a.search_vector @@ query
	`

	// Build WHERE clause for filters
	whereClauses := []string{}
	args := []interface{}{query.Query}

	if query.Category != "" {
		whereClauses = append(whereClauses, "a.normalized_metadata->>'category' = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, query.Category)
	}

	if query.Newsgroup != "" {
		whereClauses = append(whereClauses, "n.name = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, query.Newsgroup)
	}

	if query.Author != "" {
		whereClauses = append(whereClauses, "a.author ILIKE $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, "%"+query.Author+"%")
	}

	if query.MinSize != nil {
		whereClauses = append(whereClauses, "a.article_size >= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MinSize)
	}

	if query.MaxSize != nil {
		whereClauses = append(whereClauses, "a.article_size <= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MaxSize)
	}

	if query.MinDate != nil {
		whereClauses = append(whereClauses, "a.posted_at >= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MinDate)
	}

	if query.MaxDate != nil {
		whereClauses = append(whereClauses, "a.posted_at <= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MaxDate)
	}

	if len(whereClauses) > 0 {
		baseSQL += " AND " + strings.Join(whereClauses, " AND ")
	}

	// Determine sorting
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "rank"
	}
	sortOrder := query.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	if sortBy == "rank" {
		baseSQL += fmt.Sprintf(" ORDER BY rank %s", sortOrder)
	} else {
		baseSQL += fmt.Sprintf(" ORDER BY a.%s %s", sortBy, sortOrder)
	}

	// Count total
	countSQL := strings.Replace(baseSQL,
		"SELECT a.id, a.message_id, a.newsgroup_id, n.name as newsgroup_name, a.subject, a.author, a.posted_at, a.article_size, a.normalized_metadata, a.created_at, ts_rank_cd(a.search_vector, query) AS rank FROM articles a",
		"SELECT count(*) FROM articles a",
		1,
	)
	countSQL = strings.Split(countSQL, "ORDER BY")[0]

	var total int64
	err := r.db.NewRaw(countSQL, args...).Scan(ctx, &total)
	if err != nil {
		return nil, fmt.Errorf("count search results: %w", err)
	}

	// Apply pagination
	pagedSQL := baseSQL + fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.PerPage, pagination.Offset())

	var articles []models.Article
	if err := r.db.NewRaw(pagedSQL, args...).Scan(ctx, &articles); err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}

	return &PaginatedResult{
		Data:       articles,
		Total:      total,
		Page:       int(pagination.Page),
		PerPage:    int(pagination.PerPage),
		TotalPages: totalPages(total, int64(pagination.PerPage)),
	}, nil
}

// SearchSpotnetPosts searches spotnet posts using PostgreSQL full-text search.
func (r *searchRepository) SearchSpotnetPosts(ctx context.Context, query models.SearchQuery) (*PaginatedResult, error) {
	pagination := models.NewPagination(query.Page, query.PerPage)

	baseSQL := `
		SELECT sp.id, sp.external_id, sp.title, sp.description, sp.author,
		       sp.category_id, c.name as category_name, sp.posted_at,
		       sp.source_id, src.name as source_name,
		       sp.normalized_metadata, sp.created_at,
		       ts_rank_cd(sp.search_vector, query) AS rank
		FROM spotnet_posts sp
		LEFT JOIN categories c ON sp.category_id = c.id
		LEFT JOIN spotnet_sources src ON sp.source_id = src.id,
		     websearch_to_tsquery('simple', $1) AS query
		WHERE sp.search_vector @@ query
	`

	whereClauses := []string{}
	args := []interface{}{query.Query}

	if query.Category != "" {
		whereClauses = append(whereClauses, "c.slug = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, query.Category)
	}

	if query.SourceType != "" {
		whereClauses = append(whereClauses, "src.source_type = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, query.SourceType)
	}

	if query.Author != "" {
		whereClauses = append(whereClauses, "sp.author ILIKE $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, "%"+query.Author+"%")
	}

	if query.MinSize != nil {
		whereClauses = append(whereClauses, "sp.raw_metadata->>'size'::text >= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MinSize)
	}

	if query.MaxSize != nil {
		whereClauses = append(whereClauses, "sp.raw_metadata->>'size'::text <= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MaxSize)
	}

	if query.MinDate != nil {
		whereClauses = append(whereClauses, "sp.posted_at >= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MinDate)
	}

	if query.MaxDate != nil {
		whereClauses = append(whereClauses, "sp.posted_at <= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *query.MaxDate)
	}

	if len(whereClauses) > 0 {
		baseSQL += " AND " + strings.Join(whereClauses, " AND ")
	}

	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "rank"
	}
	sortOrder := query.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	if sortBy == "rank" {
		baseSQL += fmt.Sprintf(" ORDER BY rank %s", sortOrder)
	} else {
		baseSQL += fmt.Sprintf(" ORDER BY sp.%s %s", sortBy, sortOrder)
	}

	countSQL := strings.Replace(baseSQL,
		"SELECT sp.id, sp.external_id, sp.title, sp.description, sp.author, sp.category_id, c.name as category_name, sp.posted_at, sp.source_id, src.name as source_name, sp.normalized_metadata, sp.created_at, ts_rank_cd(sp.search_vector, query) AS rank FROM spotnet_posts sp",
		"SELECT count(*) FROM spotnet_posts sp",
		1,
	)
	countSQL = strings.Split(countSQL, "ORDER BY")[0]

	var total int64
	err := r.db.NewRaw(countSQL, args...).Scan(ctx, &total)
	if err != nil {
		return nil, fmt.Errorf("count search results: %w", err)
	}

	pagedSQL := baseSQL + fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.PerPage, pagination.Offset())

	var posts []models.SpotnetPost
	if err := r.db.NewRaw(pagedSQL, args...).Scan(ctx, &posts); err != nil {
		return nil, fmt.Errorf("search spotnet posts: %w", err)
	}

	return &PaginatedResult{
		Data:       posts,
		Total:      total,
		Page:       int(pagination.Page),
		PerPage:    int(pagination.PerPage),
		TotalPages: totalPages(total, int64(pagination.PerPage)),
	}, nil
}

// SearchAll searches both articles and spotnet posts.
func (r *searchRepository) SearchAll(ctx context.Context, query models.SearchQuery) (*PaginatedResult, error) {
	// For now, delegate to article search
	return r.SearchArticles(ctx, query)
}

// SearchArticlesByCategory retrieves articles filtered by category.
func (r *searchRepository) SearchArticlesByCategory(ctx context.Context, category string, pagination Pagination) (*PaginatedResult, error) {
	var articles []models.Article
	query := r.db.NewSelect().Model(&articles).
		Where("normalized_metadata->>'category' = ?", category)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	query = query.
		Offset(int64((pagination.Page - 1) * pagination.PerPage)).
		Limit(int64(pagination.PerPage)).
		Order("posted_at DESC")

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return &PaginatedResult{
		Data:       articles,
		Total:      total,
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		TotalPages: totalPages(total, int64(pagination.PerPage)),
	}, nil
}

// GetFacets returns facet counts for search results.
func (r *searchRepository) GetFacets(ctx context.Context, query models.SearchQuery) (map[string]int64, error) {
	facets := make(map[string]int64)

	// Count by newsgroup
	type newsgroupCount struct {
		Name  string `bun:"name"`
		Count int64  `bun:"count"`
	}
	var newsgroupCounts []newsgroupCount
	err := r.db.NewSelect().
		Column("n.name, count(*) as count").
		Table("articles a").
		Join("JOIN newsgroups n ON a.newsgroup_id = n.id").
		Where("a.search_vector @@ websearch_to_tsquery('simple', $1)", query.Query).
		Group("n.name").
		Scan(ctx, &newsgroupCounts)
	if err != nil {
		facets["newsgroup_error"] = 0
	} else {
		for _, nc := range newsgroupCounts {
			facets["newsgroup:"+nc.Name] = nc.Count
		}
	}

	// Count by date range
	type dateRange struct {
		MinDate time.Time `bun:"min_date"`
		MaxDate time.Time `bun:"max_date"`
	}
	var dr dateRange
	err = r.db.NewSelect().
		Column("min(posted_at) as min_date, max(posted_at) as max_date").
		Model(&models.Article{}).
		Where("search_vector @@ websearch_to_tsquery('simple', $1)", query.Query).
		Scan(ctx, &dr)
	if err == nil {
		facets["date_min"] = dr.MinDate.Unix()
		facets["date_max"] = dr.MaxDate.Unix()
	}

	return facets, nil
}

// totalPages calculates total pages from total items and perPage.
func totalPages(total, perPage int64) int {
	pages := total / perPage
	if total%perPage > 0 {
		pages++
	}
	return int(pages)
}

// Ensure interface compliance
var _ SearchRepository = (*searchRepository)(nil)
