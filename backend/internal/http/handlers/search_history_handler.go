package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/http/middleware"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// SearchHistoryHandler handles search history endpoints.
type SearchHistoryHandler struct {
	DB   *bun.DB
	Log  interface{}
}

// CreateSearchHistoryRequest represents a request to save search history.
type CreateSearchHistoryRequest struct {
	SearchQuery string                 `json:"search_query" validate:"required"`
	Filters     map[string]interface{} `json:"filters"`
}

// List handles listing search history.
func (h *SearchHistoryHandler) List(c echo.Context) error {
	userID, ok := middleware.GetUserID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var history []models.UserSearchHistory
	err := h.DB.NewSelect().
		Model(&history).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Scan(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list search history: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": history,
	})
}

// Create handles creating a search history entry.
func (h *SearchHistoryHandler) Create(c echo.Context) error {
	userID, ok := middleware.GetUserID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req CreateSearchHistoryRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	var filtersJSON []byte
	if req.Filters != nil {
		filtersJSON, _ = json.Marshal(req.Filters)
	}

	historyEntry := models.UserSearchHistory{
		UserID:      userID,
		SearchQuery: req.SearchQuery,
		Filters:     string(filtersJSON),
		CreatedAt:   time.Now().UTC(),
	}

	_, err := h.DB.NewInsert().Model(&historyEntry).Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to save search history: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "search history saved",
		"entry":   historyEntry,
	})
}

// Clear handles clearing search history.
func (h *SearchHistoryHandler) Clear(c echo.Context) error {
	userID, ok := middleware.GetUserID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	_, err := h.DB.NewDelete().
		Model(&models.UserSearchHistory{}).
		Where("user_id = ?", userID).
		Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to clear search history: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "search history cleared",
	})
}