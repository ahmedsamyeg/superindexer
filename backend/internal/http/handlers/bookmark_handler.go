package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/http/middleware"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// BookmarkHandler handles bookmark endpoints.
type BookmarkHandler struct {
	DB   *bun.DB
	Log  interface{}
}

// CreateBookmarkRequest represents a request to create a bookmark.
type CreateBookmarkRequest struct {
	ItemID   int64  `json:"item_id" validate:"required"`
	ItemType string `json:"item_type" validate:"required"`
}

// List handles listing bookmarks.
func (h *BookmarkHandler) List(c echo.Context) error {
	userID, ok := middleware.GetUserID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var bookmarks []models.Bookmark
	err := h.DB.NewSelect().
		Model(&bookmarks).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list bookmarks: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": bookmarks,
	})
}

// Create handles creating a bookmark.
func (h *BookmarkHandler) Create(c echo.Context) error {
	userID, ok := middleware.GetUserID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req CreateBookmarkRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	bookmark := models.Bookmark{
		UserID:   userID,
		ItemID:   req.ItemID,
		ItemType: req.ItemType,
	}

	_, err := h.DB.NewInsert().Model(&bookmark).
		On("CONFLICT (user_id, item_id, item_type) DO NOTHING").
		Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to create bookmark: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"bookmark": bookmark,
	})
}

// Delete handles deleting a bookmark.
func (h *BookmarkHandler) Delete(c echo.Context) error {
	bookmarkID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bookmark ID")
	}

	result, err := h.DB.NewDelete().
		Model(&models.Bookmark{}).
		Where("id = ?", bookmarkID).
		Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete bookmark: "+err.Error())
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "bookmark not found")
	}

	return c.NoContent(http.StatusNoContent)
}