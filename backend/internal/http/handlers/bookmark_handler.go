package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// BookmarkHandler handles bookmark endpoints.
type BookmarkHandler struct {
	DB interface{}
	Log interface{}
}

// List handles listing bookmarks.
func (h *BookmarkHandler) List(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": []interface{}{},
	})
}

// Create handles creating a bookmark.
func (h *BookmarkHandler) Create(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"bookmark": map[string]interface{}{
			"id":       1,
			"item_id":  1,
			"item_type": "article",
		},
	})
}

// Delete handles deleting a bookmark.
func (h *BookmarkHandler) Delete(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
