package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// NewsgroupHandler handles newsgroup management endpoints.
type NewsgroupHandler struct {
	DB   *bun.DB
	Log  interface{}
	Repo repositories.NewsgroupRepository
}

// CreateNewsgroupRequest represents a request to create a newsgroup.
type CreateNewsgroupRequest struct {
	Name             string `json:"name" validate:"required"`
	Description      string `json:"description"`
	IndexingPriority int    `json:"indexing_priority"`
}

// UpdateNewsgroupRequest represents a request to update a newsgroup.
type UpdateNewsgroupRequest struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	Enabled          *bool   `json:"enabled"`
	IndexingPriority *int    `json:"indexing_priority"`
}

// List handles listing newsgroups.
func (h *NewsgroupHandler) List(c echo.Context) error {
	groups, err := h.Repo.GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list groups: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  groups,
		"total": int64(len(groups)),
	})
}

// GetByID handles getting a newsgroup by ID.
func (h *NewsgroupHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group ID")
	}

	group, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"group": group,
	})
}

// Create handles creating a new newsgroup.
func (h *NewsgroupHandler) Create(c echo.Context) error {
	var req CreateNewsgroupRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	group := &models.Newsgroup{
		Name:             req.Name,
		Description:      req.Description,
		Enabled:          true,
		IndexingPriority: req.IndexingPriority,
	}

	if err := h.Repo.Create(c.Request().Context(), group); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to create group: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"group": group,
	})
}

// Update handles updating a newsgroup.
func (h *NewsgroupHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group ID")
	}

	group, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}

	var req UpdateNewsgroupRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.Enabled != nil {
		group.Enabled = *req.Enabled
	}
	if req.IndexingPriority != nil {
		group.IndexingPriority = *req.IndexingPriority
	}

	if err := h.Repo.Update(c.Request().Context(), group); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update group: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"group": group,
	})
}

// Delete handles deleting a newsgroup.
func (h *NewsgroupHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group ID")
	}

	if err := h.Repo.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}

	return c.NoContent(http.StatusNoContent)
}

// Discover handles newsgroup discovery.
func (h *NewsgroupHandler) Discover(c echo.Context) error {
	// In a real implementation, this would:
	// 1. Connect to enabled NNTP servers
	// 2. Fetch the list of available groups
	// 3. Compare with existing newsgroups
	// 4. Add new groups and update existing ones
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "discovery initiated",
	})
}

// ManualIndex handles manual indexing trigger.
func (h *NewsgroupHandler) ManualIndex(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group ID")
	}

	// In a real implementation, this would:
	// 1. Create an indexing job for the newsgroup
	// 2. Return the job ID for tracking
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "indexing job created",
		"group_id": id,
	})
}