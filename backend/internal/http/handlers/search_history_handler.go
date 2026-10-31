package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// SearchHistoryHandler handles search history endpoints.
type SearchHistoryHandler struct {
	DB interface{}
	Log interface{}
}

// List handles listing search history.
func (h *SearchHistoryHandler) List(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": []interface{}{},
	})
}

// Create handles creating a search history entry.
func (h *SearchHistoryHandler) Create(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "search history saved",
	})
}

// Clear handles clearing search history.
func (h *SearchHistoryHandler) Clear(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "search history cleared",
	})
}
