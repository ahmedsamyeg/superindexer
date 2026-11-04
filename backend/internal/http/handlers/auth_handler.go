package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/auth"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/middleware"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	AuthService    service.AuthServiceInterface
	DB             *bun.DB
	Log            interface{}
	JWTSecret      string
	RefreshTokens  *service.RefreshTokenService
	TokenBlacklist *service.TokenBlacklistService
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

// RefreshTokenRequest represents a token refresh request body.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a token refresh response.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
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

// Logout handles user logout and token revocation.
func (h *AuthHandler) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			claims, err := auth.ValidateToken(tokenString, auth.JWTConfig{
				Secret: []byte(h.JWTSecret),
				Issuer: "superindexer",
				Expiry: 3600,
			})
			if err == nil {
				// Blacklist the token
				tokenHash := fmt.Sprintf("%d-%d", claims.UserID, claims.ExpiresAt.Unix())
				_ = h.TokenBlacklist.BlacklistToken(c.Request().Context(), tokenHash, claims.UserID, claims.ExpiresAt.Time)
			}
		}
	}

	// Also revoke refresh token if provided
	if refreshToken := c.QueryParam("refresh_token"); refreshToken != "" {
		_ = h.RefreshTokens.RevokeRefreshToken(c.Request().Context(), refreshToken)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "logged out successfully",
	})
}

// GetMe handles getting the current user's profile.
func (h *AuthHandler) GetMe(c echo.Context) error {
	userID, ok := middleware.GetUserID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	// Fetch the user from the database
	userRepo := repositories.NewUserRepository(h.DB)
	user, err := userRepo.GetByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	// Don't expose the password hash
	return c.JSON(http.StatusOK, UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
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

	// Revoke all refresh tokens for this user (they should re-authenticate)
	userID, ok := middleware.GetUserID(c.Request().Context())
	if ok {
		_ = h.RefreshTokens.RevokeAllUserTokens(c.Request().Context(), userID)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "password reset successfully",
	})
}

// Refresh handles token refresh using a refresh token.
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshTokenRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	pair, err := h.RefreshTokens.RefreshAccessKey(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
	}

	return c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	})
}