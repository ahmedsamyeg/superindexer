package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/auth"
	"github.com/superindexer/indexer/internal/config"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
)

// AuthService handles authentication operations.
type AuthService struct {
	userRepo repositories.UserRepository
	cfg      config.AuthConfig
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repositories.UserRepository, cfg config.AuthConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// LoginResult holds the result of a login attempt.
type LoginResult struct {
	Token  string       `json:"token"`
	User   *models.User `json:"user"`
}

// Login authenticates a user by username and password.
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	// Try to find user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	err = auth.VerifyPassword(password, user.PasswordHash, auth.DefaultPasswordConfig)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Role, user.Username, auth.JWTConfig{
		Secret: []byte(s.cfg.JWTSecret),
		Issuer: "superindexer",
		Expiry: int(s.cfg.JWTExpiry.Seconds()),
	})
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResult{
		Token: token,
		User:  user,
	}, nil
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*models.User, error) {
	// Check if username already exists
	existing, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil && existing != nil {
		return nil, errors.New("username already taken")
	}

	// Check if email already exists
	existingEmail, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingEmail != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	passwordHash, err := auth.GeneratePasswordHash(password, auth.DefaultPasswordConfig)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "user",
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// ValidateToken validates a JWT token and returns user info.
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*auth.TokenClaims, error) {
	return auth.ValidateToken(tokenString, auth.JWTConfig{
		Secret: []byte(s.cfg.JWTSecret),
		Issuer: "superindexer",
		Expiry: int(s.cfg.JWTExpiry.Seconds()),
	})
}

// RequestPasswordReset generates a reset token and stores it for the user.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if the email exists - just return success
		return nil
	}

	// Generate a secure password reset token (in a real system, this would be stored in DB with expiration)
	token, err := GeneratePasswordResetToken()
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	// Set token expiry (e.g., 1 hour from now)
	expiry := time.Now().Add(1 * time.Hour)

	// Store the token in database (in a real system this would be stored with an expiry)
	err = s.userRepo.UpdatePasswordResetToken(ctx, user.ID, token, expiry)
	if err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	// In a real system, we would:
	// 1. Send an email with the reset link (e.g., to: https://yoursite.com/reset?token=...)
	// 2. Log the password reset request
	// For this phase, we'll just simulate that it was sent successfully
	
	return nil
}

// ConfirmPasswordReset validates the reset token and updates the user's password.
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// Validate the token format
	if !ValidatePasswordResetToken(token) {
		return errors.New("invalid reset token")
	}

	// Find user by reset token
	user, err := s.userRepo.GetByPasswordResetToken(ctx, token)
	if err != nil {
		return errors.New("invalid reset token")
	}

	// Check if token has expired
	if user.PasswordResetExpiry == nil || user.PasswordResetExpiry.Before(time.Now()) {
		return errors.New("reset token has expired")
	}

	// Hash the new password
	passwordHash, err := auth.GeneratePasswordHash(newPassword, auth.DefaultPasswordConfig)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	// Update the user's password hash
	err = s.userRepo.UpdatePasswordHash(ctx, user.ID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Clear the reset token from the database
	err = s.userRepo.DeletePasswordResetToken(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("clear reset token: %w", err)
	}

	return nil
}

// GeneratePasswordResetToken creates a secure random token for password reset.
func GeneratePasswordResetToken() (string, error) {
	// In a production system, this would use crypto/rand to generate a secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// ValidatePasswordResetToken checks if a password reset token is valid.
func ValidatePasswordResetToken(token string) bool {
	// In a production system, this would check token expiration and validity
	// For now, just verify it's a valid base64 string
	_, err := base64.URLEncoding.DecodeString(token)
	return err == nil
}

// UserService handles user management operations.
type UserService struct {
	userRepo repositories.UserRepository
	authCfg  config.AuthConfig
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repositories.UserRepository, authCfg config.AuthConfig) *UserService {
	return &UserService{
		userRepo: userRepo,
		authCfg:  authCfg,
	}
}

// CreateUser creates a new user with the given details.
func (s *UserService) CreateUser(ctx context.Context, username, email, password, role string) (*models.User, error) {
	// Hash password
	passwordHash, err := auth.GeneratePasswordHash(password, auth.DefaultPasswordConfig)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if role == "" {
		role = "user"
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID.
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// ListUsers retrieves a paginated list of users.
func (s *UserService) ListUsers(ctx context.Context, page, perPage int) (*models.PaginatedResult, error) {
	pagination := repositories.Pagination{
		Page:    page,
		PerPage: perPage,
	}

	result, err := s.userRepo.List(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return &models.PaginatedResult{
		Data:       result.Data,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}, nil
}

// UpdateUser updates a user's fields.
func (s *UserService) UpdateUser(ctx context.Context, id int64, username, email, password, role *string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if username != nil {
		user.Username = *username
	}
	if email != nil {
		user.Email = *email
	}
	if password != nil {
		hash, err := auth.GeneratePasswordHash(*password, auth.DefaultPasswordConfig)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		user.PasswordHash = hash
	}
	if role != nil {
		user.Role = *role
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user by ID.
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.userRepo.Delete(ctx, id)
}

// ToggleUserActive toggles a user's active status.
func (s *UserService) ToggleUserActive(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.IsActive = !user.IsActive

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("toggle user active: %w", err)
	}

	return user, nil
}