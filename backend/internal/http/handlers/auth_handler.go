package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/auth"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/superindexer/indexer/internal/http/validator"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	AuthService *service.AuthService
	Log         interface{}
}

// LoginRequest represents a login request body.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents a registration request body.
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// PasswordResetRequest represents a password reset request body.
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetConfirmRequest represents a password reset confirmation request body.
type PasswordResetConfirmRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	Token  string `json:"token"`
	User   UserResponse `json:"user"`
}

// UserResponse represents a user response.
type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// Login handles user login.
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	loginResult, err := h.AuthService.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: loginResult.Token,
		User: UserResponse{
			ID:       loginResult.User.ID,
			Username: loginResult.User.Username,
			Email:    loginResult.User.Email,
			Role:     loginResult.User.Role,
			IsActive: loginResult.User.IsActive,
		},
	})
}

// Register handles user registration.
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.AuthService.Register(c.Request().Context(), req.Username, req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	})
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c echo.Context) error {
	// JWT tokens are stateless, so logout is handled by client-side token removal
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "logged out successfully",
	})
}

// GetMe handles getting the current user's profile.
func (h *AuthHandler) GetMe(c echo.Context) error {
	userID, role, username, ok := auth.ExtractUserFromContext(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	// In a real implementation, we would fetch the user from DB
	// For now, we'll return placeholder data based on context
	return c.JSON(http.StatusOK, UserResponse{
		ID:       userID,
		Username: username,
		Email:    username + "@localhost", // Placeholder
		Role:     role,
		IsActive: true, // Placeholder
	})
}

// RequestPasswordReset handles password reset requests.
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var req PasswordResetRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	err := h.AuthService.RequestPasswordReset(c.Request().Context(), req.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process password reset request")
	}

	// In a real implementation, we would send an email with the reset link
	// For now, we'll just return success to indicate the process started
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "if the email exists, a reset link has been sent",
	})
}

// ConfirmPasswordReset handles password reset confirmation.
func (h *AuthHandler) ConfirmPasswordReset(c echo.Context) error {
	var req PasswordResetConfirmRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	err := h.AuthService.ConfirmPasswordReset(c.Request().Context(), req.Token, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "password reset successfully",
	})
}