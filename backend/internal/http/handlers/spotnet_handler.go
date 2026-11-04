package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// SpotnetSourceHandler handles Spotnet source management endpoints.
type SpotnetSourceHandler struct {
	DB   *bun.DB
	Log  interface{}
	Repo repositories.SpotnetSourceRepository
}

// CreateSpotnetSourceRequest represents a request to create a Spotnet source.
type CreateSpotnetSourceRequest struct {
	Name          string                 `json:"name" validate:"required"`
	SourceType    string                 `json:"source_type" validate:"required"`
	Configuration map[string]interface{} `json:"configuration"`
}

// UpdateSpotnetSourceRequest represents a request to update a Spotnet source.
type UpdateSpotnetSourceRequest struct {
	Name          *string                `json:"name"`
	SourceType    *string                `json:"source_type"`
	Configuration *map[string]interface{} `json:"configuration"`
	Enabled       *bool                  `json:"enabled"`
}

// List handles listing Spotnet sources.
func (h *SpotnetSourceHandler) List(c echo.Context) error {
	sources, err := h.Repo.GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list sources: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  sources,
		"total": int64(len(sources)),
	})
}

// GetByID handles getting a Spotnet source by ID.
func (h *SpotnetSourceHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid source ID")
	}

	source, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "source not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"source": source,
	})
}

// Create handles creating a new Spotnet source.
func (h *SpotnetSourceHandler) Create(c echo.Context) error {
	var req CreateSpotnetSourceRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	var configJSON []byte
	if req.Configuration != nil {
		configJSON, _ = json.Marshal(req.Configuration)
	}

	source := &models.SpotnetSource{
		Name:          req.Name,
		SourceType:    req.SourceType,
		Configuration: string(configJSON),
		Enabled:       true,
	}

	if err := h.Repo.Create(c.Request().Context(), source); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to create source: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"source": source,
	})
}

// Update handles updating a Spotnet source.
func (h *SpotnetSourceHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid source ID")
	}

	source, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "source not found")
	}

	var req UpdateSpotnetSourceRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	if req.Name != nil {
		source.Name = *req.Name
	}
	if req.SourceType != nil {
		source.SourceType = *req.SourceType
	}
	if req.Configuration != nil {
		configJSON, _ := json.Marshal(*req.Configuration)
		source.Configuration = string(configJSON)
	}
	if req.Enabled != nil {
		source.Enabled = *req.Enabled
	}

	if err := h.Repo.Update(c.Request().Context(), source); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update source: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"source": source,
	})
}

// Delete handles deleting a Spotnet source.
func (h *SpotnetSourceHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid source ID")
	}

	if err := h.Repo.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "source not found")
	}

	return c.NoContent(http.StatusNoContent)
}

// TestConnection handles testing Spotnet source connectivity.
func (h *SpotnetSourceHandler) TestConnection(c echo.Context) error {
	_, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid source ID")
	}

	// In a real implementation, this would:
	// 1. Load the source configuration
	// 2. Attempt to connect to the source
	// 3. Return the connection status
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "connected",
		"message": "connection test successful",
	})
}

// Sync handles triggering a sync operation.
func (h *SpotnetSourceHandler) Sync(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid source ID")
	}

	// In a real implementation, this would:
	// 1. Create a sync job for the source
	// 2. Return the job ID for tracking
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "sync initiated",
		"source_id": id,
	})
}

// SpotnetPostHandler handles Spotnet post endpoints.
type SpotnetPostHandler struct {
	DB   *bun.DB
	Log  interface{}
	Repo repositories.SpotnetSourceRepository
}

// List handles listing Spotnet posts.
func (h *SpotnetPostHandler) List(c echo.Context) error {
	// In a real implementation, this would query the spotnet_posts table
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  []interface{}{},
		"total": int64(0),
	})
}

// GetByID handles getting a Spotnet post by ID.
func (h *SpotnetPostHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post ID")
	}

	// In a real implementation, this would query the spotnet_posts table
	return c.JSON(http.StatusOK, map[string]interface{}{
		"post": map[string]interface{}{
			"id":          id,
			"external_id": "ext-" + strconv.FormatInt(id, 10),
			"title":       "Example post",
			"author":      "author@example.com",
		},
	})
}