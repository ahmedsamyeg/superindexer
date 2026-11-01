package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/superindexer/indexer/internal/http/validator"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	Log         interface{}
	UserService *service.UserService
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

	paginatedResult, err := h.UserService.ListUsers(c.Request().Context(), page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list users")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":       paginatedResult.Data,
		"total":      paginatedResult.Total,
		"page":       paginatedResult.Page,
		"per_page":   paginatedResult.PerPage,
		"total_pages": paginatedResult.TotalPages,
	})
}

// GetByID handles getting a user by ID.
func (h *UserHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	user, err := h.UserService.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// Create handles creating a new user.
func (h *UserHandler) Create(c echo.Context) error {
	var req CreateUserRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.UserService.CreateUser(c.Request().Context(), req.Username, req.Email, req.Password, req.Role)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user": user,
	})
}

// Update handles updating a user.
func (h *UserHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UpdateUserRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	var usernamePtr *string
	var emailPtr *string
	var passwordPtr *string
	var rolePtr *string

	if req.Username != "" {
		usernamePtr = &req.Username
	}
	if req.Email != "" {
		emailPtr = &req.Email
	}
	if req.Password != "" {
		passwordPtr = &req.Password
	}
	if req.Role != "" {
		rolePtr = &req.Role
	}

	user, err := h.UserService.UpdateUser(c.Request().Context(), id, usernamePtr, emailPtr, passwordPtr, rolePtr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// Delete handles deleting a user.
func (h *UserHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.UserService.DeleteUser(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.NoContent(http.StatusNoContent)
}

// ToggleActive handles toggling user active status.
func (h *UserHandler) ToggleActive(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	user, err := h.UserService.ToggleUserActive(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}