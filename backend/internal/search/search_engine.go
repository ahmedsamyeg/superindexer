package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/uptrace/bun"
)

// SearchEngine handles full-text search operations
type SearchEngine struct {
	repo repositories.SearchRepository
	DB   *bun.DB
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
	FileType    string
	ContentType string
	ProcessingStatus string
	Page        int
	PerPage     int
	SortBy      string
	SortOrder   string
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Articles     []*SearchResultItem `json:"articles"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	PerPage      int                 `json:"per_page"`
	TotalPages   int                 `json:"total_pages"`
	Facets       map[string]int64    `json:"facets,omitempty"`
	Query        string              `json:"query"`
	SortBy       string              `json:"sort_by"`
	SortOrder    string              `json:"sort_order"`
}

// SearchResultItem is a unified result that can be either an article or spotnet post
type SearchResultItem struct {
	ID           int64     `json:"id"`
	ExternalID   string    `json:"external_id,omitempty"`
	Source       string    `json:"source"`       // "article" or "spotnet"
	Subject      string    `json:"subject"`
	Title        string    `json:"title,omitempty"`
	Description  string    `json:"description,omitempty"`
	Author       string    `json:"author"`
	Newsgroup    string    `json:"newsgroup,omitempty"`
	Category     string    `json:"category"`
	PostedAt     time.Time `json:"posted_at"`
	Size         int64     `json:"size"`
	MessageID    string    `json:"message_id,omitempty"`
	SourceName   string    `json:"source_name,omitempty"`
	Rank         float64   `json:"rank"`
	RawMetadata  string    `json:"raw_metadata,omitempty"`
	NormalizedMetadata string `json:"normalized_metadata,omitempty"`
}

// Search performs a search across articles and spotnet posts, combining results with relevance ranking
func (se *SearchEngine) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// Validate and set defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 20
	}
	if query.PerPage > 100 {
		query.PerPage = 100
	}

	// If query is empty, return empty results
	if strings.TrimSpace(query.Query) == "" {
		return &SearchResult{
			Total:      0,
			Page:       query.Page,
			PerPage:    query.PerPage,
			TotalPages: 0,
			Facets:     make(map[string]int64),
		}, nil
	}

	// Search articles
	articleResult, err := se.SearchArticles(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search articles failed: %w", err)
	}

	// Search spotnet posts
	spotnetResult, err := se.SearchSpotnetPosts(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search spotnet posts failed: %w", err)
	}

	// Combine results with proper ranking
	combined := se.combineResults(articleResult, spotnetResult, query)

	// Get facets
	facets, err := se.GetFacets(ctx, query)
	if err != nil {
		// Don't fail the search if facets fail
		facets = make(map[string]int64)
	}

	return &SearchResult{
		Articles:     combined.Items,
		Total:        combined.Total,
		Page:         combined.Page,
		PerPage:      combined.PerPage,
		TotalPages:   combined.TotalPages,
		Facets:       facets,
		Query:        query.Query,
		SortBy:       query.SortBy,
		SortOrder:    query.SortOrder,
	}, nil
}

// combinedResults holds the combined search results
type combinedResults struct {
	Items      []*SearchResultItem
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

// combineResults merges article and spotnet results, applies ranking and pagination
func (se *SearchEngine) combineResults(articleResult *SearchResult, spotnetResult *SearchResult, query SearchQuery) *combinedResults {
	// Calculate total across both sources
	total := articleResult.Total + spotnetResult.Total

	// Build combined items from articles
	items := make([]*SearchResultItem, 0, len(articleResult.Articles)+len(spotnetResult.Articles))

	for _, article := range articleResult.Articles {
		item := &SearchResultItem{
			ID:                 article.ID,
			Source:             "article",
			Subject:            article.Subject,
			Author:             article.Author,
			Newsgroup:          article.Newsgroup,
			Category:           article.Category,
			PostedAt:           article.PostedAt,
			Size:               article.Size,
			MessageID:          article.MessageID,
			RawMetadata:        article.RawMetadata,
			NormalizedMetadata: article.NormalizedMetadata,
		}
		items = append(items, item)
	}

	// Build combined items from spotnet posts
	for _, post := range spotnetResult.Articles {
		item := &SearchResultItem{
			ID:                 post.ID,
			ExternalID:         post.ExternalID,
			Source:             "spotnet",
			Title:              post.Title,
			Description:        post.Description,
			Author:             post.Author,
			Category:           post.Category,
			PostedAt:           post.PostedAt,
			SourceName:         post.SourceName,
			RawMetadata:        post.RawMetadata,
			NormalizedMetadata: post.NormalizedMetadata,
		}
		items = append(items, item)
	}

	// Sort combined results by rank if sorting by relevance
	if query.SortBy == "" || query.SortBy == "rank" {
		items = sortResultsByRank(items, query.SortOrder)
	} else {
		items = sortResultsByField(items, query.SortBy, query.SortOrder)
	}

	// Apply pagination
	start := (query.Page - 1) * query.PerPage
	end := start + query.PerPage
	if start > len(items) {
		start = len(items)
	}
	if end > len(items) {
		end = len(items)
	}
	paginatedItems := items[start:end]

	totalPages := int(total) / query.PerPage
	if int(total)%query.PerPage > 0 {
		totalPages++
	}

	return &combinedResults{
		Items:      paginatedItems,
		Total:      total,
		Page:       query.Page,
		PerPage:    query.PerPage,
		TotalPages: totalPages,
	}
}

// sortResultsByRank sorts results by their relevance rank (descending)
func sortResultsByRank(items []*SearchResultItem, sortOrder string) []*SearchResultItem {
	// Simple insertion sort for small result sets
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 && compareByRank(items[j], key, sortOrder) > 0 {
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
	return items
}

// sortResultsByField sorts results by a specific field
func sortResultsByField(items []*SearchResultItem, sortBy, sortOrder string) []*SearchResultItem {
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 && compareByField(items[j], key, sortBy, sortOrder) > 0 {
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
	return items
}

// compareByRank compares two results by rank
func compareByRank(a, b *SearchResultItem, sortOrder string) int {
	if a.Rank > b.Rank {
		return -1
	}
	if a.Rank < b.Rank {
		return 1
	}
	return 0
}

// compareByField compares two results by a specific field
func compareByField(a, b *SearchResultItem, sortBy, sortOrder string) int {
	var cmp int
	switch sortBy {
	case "subject", "title":
		cmp = compareString(a.Subject, b.Subject)
	case "author":
		cmp = compareString(a.Author, b.Author)
	case "newsgroup":
		cmp = compareString(a.Newsgroup, b.Newsgroup)
	case "category":
		cmp = compareString(a.Category, b.Category)
	case "posted_at", "date":
		if a.PostedAt.Before(b.PostedAt) {
			cmp = -1
		} else if a.PostedAt.After(b.PostedAt) {
			cmp = 1
		}
	case "size":
		cmp = compareInt64(a.Size, b.Size)
	default:
		cmp = 0
	}

	if sortOrder == "ASC" {
		return -cmp
	}
	return cmp
}

func compareString(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareInt64(a, b int64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// extractCategory extracts category from normalized metadata JSON
func extractCategory(metadata string) string {
	if metadata == "" {
		return ""
	}
	// Simple JSON parsing to extract category
	idx := strings.Index(metadata, `"category"`)
	if idx == -1 {
		return ""
	}
	// Find the value after "category"
	valStart := strings.Index(metadata[idx:], `"`)
	if valStart == -1 {
		return ""
	}
	valStart += idx + 1
	valEnd := strings.Index(metadata[valStart:], `"`)
	if valEnd == -1 {
		return ""
	}
	return metadata[valStart : valStart+valEnd]
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
		Query:              query.Query,
		Category:           query.Category,
		Newsgroup:          query.Newsgroup,
		SourceType:         query.SourceType,
		Author:             query.Author,
		MinSize:            query.MinSize,
		MaxSize:            query.MaxSize,
		MinDate:            query.MinDate,
		MaxDate:            query.MaxDate,
		SortBy:             query.SortBy,
		SortOrder:          query.SortOrder,
		ProcessingStatus:   query.ProcessingStatus,
	}

	result, err := se.repo.SearchArticles(ctx, dbQuery)
	if err != nil {
		return nil, fmt.Errorf("search articles failed: %w", err)
	}

	// Convert to search result format
	articles := result.Data.([]*models.Article)
	searchResult := &SearchResult{
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}

	// Convert articles to SearchResultItems with rank
	for _, article := range articles {
		item := &SearchResultItem{
			ID:                 article.ID,
			Source:             "article",
			Subject:            article.Subject,
			Author:             article.Author,
			Newsgroup:          article.NewsgroupName,
			Category:           extractCategory(article.NormalizedMetadata),
			PostedAt:           article.PostedAt,
			Size:               article.ArticleSize,
			MessageID:          article.MessageID,
			RawMetadata:        article.RawMetadata,
			NormalizedMetadata: article.NormalizedMetadata,
		}
		searchResult.Articles = append(searchResult.Articles, item)
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
		SortBy:      query.SortBy,
		SortOrder:   query.SortOrder,
	}

	result, err := se.repo.SearchSpotnetPosts(ctx, dbQuery)
	if err != nil {
		return nil, fmt.Errorf("search spotnet posts failed: %w", err)
	}

	// Convert to search result format
	posts := result.Data.([]*models.SpotnetPost)
	searchResult := &SearchResult{
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}

	// Convert posts to SearchResultItems with rank
	for _, post := range posts {
		item := &SearchResultItem{
			ID:                 post.ID,
			ExternalID:         post.ExternalID,
			Source:             "spotnet",
			Title:              post.Title,
			Description:        post.Description,
			Author:             post.Author,
			Category:           post.CategoryName,
			PostedAt:           post.PostedAt,
			SourceName:         post.SourceName,
			RawMetadata:        post.RawMetadata,
			NormalizedMetadata: post.NormalizedMetadata,
		}
		searchResult.Articles = append(searchResult.Articles, item)
	}

	return searchResult, nil
}

// SearchAll combines results from articles and spotnet posts with relevance ranking
func (se *SearchEngine) SearchAll(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// Delegate to Search which already handles combining
	return se.Search(ctx, query)
}

// GetFacets retrieves facet information for search filters
func (se *SearchEngine) GetFacets(ctx context.Context, query SearchQuery) (map[string]int64, error) {
	facets := make(map[string]int64)

	// Use the repository to get facets
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

	repoFacets, err := se.repo.GetFacets(ctx, dbQuery)
	if err != nil {
		return facets, nil // Return empty facets on error
	}

	// Merge repo facets into our facets map
	for k, v := range repoFacets {
		facets[k] = v
	}

	// Add additional facets based on the search context

	// Source type facets (article vs spotnet)
	if query.Query != "" {
		// Count articles matching the query
		var articleCount int64
		err := se.DB.NewRaw(`
			SELECT count(*) FROM articles a, websearch_to_tsquery('simple', $1) AS query
			WHERE a.search_vector @@ query
		`, query.Query).Scan(ctx, &articleCount)
		if err == nil {
			facets["source:article"] = articleCount
		}

		// Count spotnet posts matching the query
		var spotnetCount int64
		err = se.DB.NewRaw(`
			SELECT count(*) FROM spotnet_posts sp, websearch_to_tsquery('simple', $1) AS query
			WHERE sp.search_vector @@ query
		`, query.Query).Scan(ctx, &spotnetCount)
		if err == nil {
			facets["source:spotnet"] = spotnetCount
		}

		// Date range facet
		type dateRange struct {
			MinDate time.Time `bun:"min_date"`
			MaxDate time.Time `bun:"max_date"`
		}
		var dr dateRange
		err = se.DB.NewRaw(`
			SELECT min(posted_at) as min_date, max(posted_at) as max_date
			FROM articles a, websearch_to_tsquery('simple', $1) AS query
			WHERE a.search_vector @@ query
		`, query.Query).Scan(ctx, &dr)
		if err == nil {
			facets["date_min"] = dr.MinDate.Unix()
			facets["date_max"] = dr.MaxDate.Unix()
		}
	}

	return facets, nil
}

// CreateSearchVector generates a PostgreSQL tsvector for search
func (se *SearchEngine) CreateSearchVector(ctx context.Context, content string) (string, error) {
	// Use PostgreSQL's to_tsvector with unaccent for better search
	// The format is: to_tsvector('simple', to_tsquery('simple', content))
	// We use simple lexicon for consistent behavior

	// Escape single quotes for SQL
	escaped := strings.ReplaceAll(content, "'", "''")

	// Return the tsvector expression
	return fmt.Sprintf("to_tsvector('simple', unaccent(%s))", "'"+escaped+"'"), nil
}

// CreateSearchQuery generates a PostgreSQL tsquery from a search string
func (se *SearchEngine) CreateSearchQuery(ctx context.Context, query string) (string, error) {
	// Use websearch_to_tsquery which handles natural language search
	escaped := strings.ReplaceAll(query, "'", "''")
	return fmt.Sprintf("websearch_to_tsquery('simple', '%s')", escaped), nil
}

// GetSearchSuggestion returns search suggestions based on partial query
func (se *SearchEngine) GetSearchSuggestion(ctx context.Context, partial string, limit int) ([]string, error) {
	if partial == "" || limit <= 0 {
		return []string{}, nil
	}

	if limit > 20 {
		limit = 20
	}

	var suggestions []string
	err := se.DB.NewRaw(`
		SELECT DISTINCT author FROM articles
		WHERE author ILIKE $1
		ORDER BY author
		LIMIT $2
	`, "%"+partial+"%", limit).Scan(ctx, &suggestions)

	if err != nil {
		return []string{}, nil
	}

	return suggestions, nil
}
