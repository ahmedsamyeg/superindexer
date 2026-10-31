package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	DB interface{}
	Log interface{}
}

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"`
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// List handles listing users.
func (h *UserHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": []interface{}{},
		"total": 0,
		"page": page,
		"per_page": perPage,
		"total_pages": 0,
	})
}

// GetByID handles getting a user by ID.
func (h *UserHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":       id,
			"username": "admin",
			"email":    "admin@localhost",
			"role":     "admin",
			"is_active": true,
		},
	})
}

// Create handles creating a new user.
func (h *UserHandler) Create(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":       1,
			"username": req.Username,
			"email":    req.Email,
			"role":     req.Role,
		},
	})
}

// Update handles updating a user.
func (h *UserHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":       id,
			"username": req.Username,
			"email":    req.Email,
			"role":     req.Role,
		},
	})
}

// Delete handles deleting a user.
func (h *UserHandler) Delete(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// ToggleActive handles toggling user active status.
func (h *UserHandler) ToggleActive(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "user status toggled",
	})
}
