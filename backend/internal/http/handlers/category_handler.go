package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// CategoryHandler handles category management endpoints.
type CategoryHandler struct {
	DB   *bun.DB
	Log  interface{}
	Repo repositories.CategoryRepository
}

// CreateCategoryRequest represents a request to create a category.
type CreateCategoryRequest struct {
	Name     string `json:"name" validate:"required"`
	ParentID *int64 `json:"parent_id"`
}

// UpdateCategoryRequest represents a request to update a category.
type UpdateCategoryRequest struct {
	Name     *string `json:"name"`
	ParentID *int64  `json:"parent_id"`
}

// List handles listing categories.
func (h *CategoryHandler) List(c echo.Context) error {
	categories, err := h.Repo.GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list categories: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  categories,
		"total": int64(len(categories)),
	})
}

// GetByID handles getting a category by ID.
func (h *CategoryHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid category ID")
	}

	category, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "category not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"category": category,
	})
}

// Create handles creating a new category.
func (h *CategoryHandler) Create(c echo.Context) error {
	var req CreateCategoryRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Generate slug from name
	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))

	category := &models.Category{
		Name:     req.Name,
		Slug:     slug,
		ParentID: req.ParentID,
	}

	if err := h.Repo.Create(c.Request().Context(), category); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to create category: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"category": category,
	})
}

// Update handles updating a category.
func (h *CategoryHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid category ID")
	}

	category, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "category not found")
	}

	var req UpdateCategoryRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	if req.Name != nil {
		category.Name = *req.Name
		category.Slug = strings.ToLower(strings.ReplaceAll(*req.Name, " ", "-"))
	}
	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}

	if err := h.Repo.Update(c.Request().Context(), category); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update category: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"category": category,
	})
}

// Delete handles deleting a category.
func (h *CategoryHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid category ID")
	}

	if err := h.Repo.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "category not found")
	}

	return c.NoContent(http.StatusNoContent)
}