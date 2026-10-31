package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// SearchHandler handles search endpoints.
type SearchHandler struct {
	DB interface{}
	Log interface{}
}

// Search handles the general search endpoint.
func (h *SearchHandler) Search(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data":       []interface{}{},
			"total":      0,
			"page":       1,
			"per_page":   20,
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": []interface{}{},
		"total": 0,
		"page": page,
		"per_page": perPage,
		"total_pages": 0,
		"query": map[string]interface{}{
			"q":       q,
			"category": category,
			"newsgroup": newsgroup,
			"author": author,
			"min_date": minDate,
			"max_date": maxDate,
			"sort_by":  sortBy,
			"sort_order": sortOrder,
		},
	})
}

// SearchArticles handles article-specific search.
func (h *SearchHandler) SearchArticles(c echo.Context) error {
	return h.Search(c)
}

// SearchSpotnet handles spotnet post search.
func (h *SearchHandler) SearchSpotnet(c echo.Context) error {
	return h.Search(c)
}

// GetFacets handles getting search facets.
func (h *SearchHandler) GetFacets(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"facets": map[string]interface{}{
			"newsgroups": []interface{}{},
			"categories": []interface{}{},
			"date_range": map[string]interface{}{
				"min": time.Now().AddDate(0, -1, 0).Unix(),
				"max": time.Now().Unix(),
			},
		},
	})
}
