package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/http/middleware"
	"github.com/superindexer/indexer/internal/search"
)

// SearchHandler handles search endpoints.
type SearchHandler struct {
	Log          interface{}
	SearchEngine *search.SearchEngine
}

// SearchRequest represents a search request body.
type SearchRequest struct {
	Query     string     `json:"query" validate:"required"`
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
	Category  string     `json:"category"`
	Newsgroup string     `json:"newsgroup"`
	Author    string     `json:"author"`
	MinSize   *int64     `json:"min_size"`
	MaxSize   *int64     `json:"max_size"`
	MinDate   *time.Time `json:"min_date"`
	MaxDate   *time.Time `json:"max_date"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order"`
}

// Search handles the general search endpoint.
func (h *SearchHandler) Search(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data":        []interface{}{},
			"total":       int64(0),
			"page":        1,
			"per_page":    20,
			"total_pages": 0,
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Parse optional filters
	var minDate, maxDate *time.Time
	if md := c.QueryParam("min_date"); md != "" {
		t, err := time.Parse(time.RFC3339, md)
		if err == nil {
			minDate = &t
		}
	}
	if mx := c.QueryParam("max_date"); mx != "" {
		t, err := time.Parse(time.RFC3339, mx)
		if err == nil {
			maxDate = &t
		}
	}

	category := c.QueryParam("category")
	newsgroup := c.QueryParam("newsgroup")
	author := c.QueryParam("author")
	sortBy := c.QueryParam("sort_by")
	sortOrder := c.QueryParam("sort_order")

	// Use the search engine
	result, err := h.SearchEngine.Search(c.Request().Context(), search.SearchQuery{
		Query:     q,
		Category:  category,
		Newsgroup: newsgroup,
		Author:    author,
		MinDate:   minDate,
		MaxDate:   maxDate,
		Page:      page,
		PerPage:   perPage,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "search failed: "+err.Error())
	}

	// Save to search history if user is authenticated
	if userID, ok := middleware.GetUserID(c.Request().Context()); ok {
		h.saveSearchHistory(c, userID, q)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        result.Articles,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
		"facets":      result.Facets,
		"query": map[string]interface{}{
			"q":         q,
			"category":  category,
			"newsgroup": newsgroup,
			"author":    author,
			"sort_by":   sortBy,
			"sort_order": sortOrder,
		},
	})
}

// SearchArticles handles searching only articles.
func (h *SearchHandler) SearchArticles(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data":        []interface{}{},
			"total":       int64(0),
			"page":        1,
			"per_page":    20,
			"total_pages": 0,
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	result, err := h.SearchEngine.SearchArticles(c.Request().Context(), search.SearchQuery{
		Query:   q,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "search failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        result.Articles,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
		"source":      "articles",
	})
}

// SearchSpotnet handles searching only Spotnet posts.
func (h *SearchHandler) SearchSpotnet(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data":        []interface{}{},
			"total":       int64(0),
			"page":        1,
			"per_page":    20,
			"total_pages": 0,
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	result, err := h.SearchEngine.SearchSpotnetPosts(c.Request().Context(), search.SearchQuery{
		Query:   q,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "search failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        result.Articles,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
		"source":      "spotnet",
	})
}

// GetFacets handles getting search facets.
func (h *SearchHandler) GetFacets(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"facets": map[string]int64{},
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	facets, err := h.SearchEngine.GetFacets(c.Request().Context(), search.SearchQuery{
		Query:   q,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get facets: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"facets": facets,
	})
}

// GetCategories handles getting available categories.
func (h *SearchHandler) GetCategories(c echo.Context) error {
	var categories []models.Category
	err := h.SearchEngine.DB.NewSelect().
		Model(&categories).
		Order("name ASC").
		Scan(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get categories: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": categories,
	})
}

// GetSuggestions handles search suggestions.
func (h *SearchHandler) GetSuggestions(c echo.Context) error {
	partial := c.QueryParam("q")
	if partial == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"suggestions": []string{},
		})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	suggestions, err := h.SearchEngine.GetSearchSuggestion(c.Request().Context(), partial, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get suggestions: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"suggestions": suggestions,
	})
}

// saveSearchHistory saves a search query to the user's history.
func (h *SearchHandler) saveSearchHistory(c echo.Context, userID int64, query string) {
	historyEntry := models.UserSearchHistory{
		UserID:      userID,
		SearchQuery: query,
		CreatedAt:   time.Now().UTC(),
	}

	// Use a goroutine to avoid blocking the response
	go func() {
		_, err := h.SearchEngine.DB.NewInsert().Model(&historyEntry).Exec(c.Request().Context())
		if err != nil {
			// Log the error but don't fail the request
			if l, ok := h.Log.(*interface{}); ok {
				_ = l
			}
		}
	}()
}
