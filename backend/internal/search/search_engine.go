package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
)

// SearchEngine handles full-text search operations
type SearchEngine struct {
	repo repositories.SearchRepository
}

// NewSearchEngine creates a new search engine instance
func NewSearchEngine(repo repositories.SearchRepository) *SearchEngine {
	return &SearchEngine{
		repo: repo,
	}
}

// SearchQuery represents a search query with various filters and options
type SearchQuery struct {
	Query       string
	Category    string
	Newsgroup   string
	SourceType  string
	Author      string
	MinSize     *int64
	MaxSize     *int64
	MinDate     *time.Time
	MaxDate     *time.Time
	Page        int
	PerPage     int
	SortBy      string
	SortOrder   string
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Articles     []*models.Article     `json:"articles"`
	SpotnetPosts []*models.SpotnetPost `json:"spotnet_posts"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	PerPage      int                   `json:"per_page"`
	TotalPages   int                   `json:"total_pages"`
}

// Search performs a search across articles and spotnet posts
func (se *SearchEngine) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// In a real implementation:
	// 1. Validate query parameters
	// 2. Execute search across both article and spotnet post tables
	// 3. Combine and sort results by relevance
	// 4. Return paginated results
	
	// For now, return a simplified result structure
	return &SearchResult{
		Articles:     []*models.Article{},
		SpotnetPosts: []*models.SpotnetPost{},
		Total:        0,
		Page:         query.Page,
		PerPage:      query.PerPage,
		TotalPages:   0,
	}, nil
}

// SearchArticles searches only articles using PostgreSQL FTS
func (se *SearchEngine) SearchArticles(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// Validate input parameters
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 20
	}

	// Convert to database query model
	dbQuery := models.SearchQuery{
		Query:       query.Query,
		Category:    query.Category,
		Newsgroup:   query.Newsgroup,
		SourceType:  query.SourceType,
		Author:      query.Author,
		MinSize:     query.MinSize,
		MaxSize:     query.MaxSize,
		MinDate:     query.MinDate,
		MaxDate:     query.MaxDate,
		Page:        query.Page,
		PerPage:     query.PerPage,
		SortBy:      query.SortBy,
		SortOrder:   query.SortOrder,
	}

	result, err := se.repo.SearchArticles(ctx, dbQuery)
	if err != nil {
		return nil, fmt.Errorf("search articles failed: %w", err)
	}

	// Convert to search result format
	searchResult := &SearchResult{
		Articles:   result.Data.([]*models.Article),
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}

	return searchResult, nil
}

// SearchSpotnetPosts searches only Spotnet posts using PostgreSQL FTS
func (se *SearchEngine) SearchSpotnetPosts(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// Validate input parameters
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 20
	}

	// Convert to database query model
	dbQuery := models.SearchQuery{
		Query:       query.Query,
		Category:    query.Category,
		Newsgroup:   query.Newsgroup,
		SourceType:  query.SourceType,
		Author:      query.Author,
		MinSize:     query.MinSize,
		MaxSize:     query.MaxSize,
		MinDate:     query.MinDate,
		MaxDate:     query.MaxDate,
		Page:        query.Page,
		PerPage:     query.PerPage,
		SortBy:      query.SortBy,
		SortOrder:   query.SortOrder,
	}

	result, err := se.repo.SearchSpotnetPosts(ctx, dbQuery)
	if err != nil {
		return nil, fmt.Errorf("search spotnet posts failed: %w", err)
	}

	// Convert to search result format
	searchResult := &SearchResult{
		SpotnetPosts: result.Data.([]*models.SpotnetPost),
		Total:        result.Total,
		Page:         result.Page,
		PerPage:      result.PerPage,
		TotalPages:   result.TotalPages,
	}

	return searchResult, nil
}

// SearchAll combines results from articles and spotnet posts with relevance ranking
func (se *SearchEngine) SearchAll(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// In a real implementation:
	// 1. Execute search on both article and spotnet post tables
	// 2. Combine results with proper relevance scoring
	// 3. Sort by combined rank or other criteria
	// 4. Apply pagination
	
	// For now return placeholder result
	return &SearchResult{
		Articles:     []*models.Article{},
		SpotnetPosts: []*models.SpotnetPost{},
		Total:        0,
		Page:         query.Page,
		PerPage:      query.PerPage,
		TotalPages:   0,
	}, nil
}

// GetFacets retrieves facet information for search filters
func (se *SearchEngine) GetFacets(ctx context.Context, query SearchQuery) (map[string]int64, error) {
	// In a real implementation:
	// 1. Query database to get counts for various filter values
	// 2. Return map of facet names to count values
	
	dbQuery := models.SearchQuery{
		Query:       query.Query,
		Category:    query.Category,
		Newsgroup:   query.Newsgroup,
		SourceType:  query.SourceType,
		Author:      query.Author,
		Page:        query.Page,
		PerPage:     query.PerPage,
		SortBy:      query.SortBy,
		SortOrder:   query.SortOrder,
	}

	facets, err := se.repo.GetFacets(ctx, dbQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get facets: %w", err)
	}

	return facets, nil
}

// CreateSearchVector generates a PostgreSQL tsvector for search
func (se *SearchEngine) CreateSearchVector(ctx context.Context, content string) (string, error) {
	// In a real implementation:
	// 1. Parse content into appropriate fields for search
	// 2. Create properly formatted tsvector string
	// 3. Handle special cases like date formatting
	
	return fmt.Sprintf("'%s'", strings.ReplaceAll(content, "'", "''")), nil
}