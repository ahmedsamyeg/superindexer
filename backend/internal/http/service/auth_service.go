package service

import (
	"context"
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
		return nil, fmt.Errorf("toggle active: %w", err)
	}

	return user, nil
}

// ListUsers returns a paginated list of users.
func (s *UserService) ListUsers(ctx context.Context, page, perPage int) (*repositories.PaginatedResult, error) {
	return s.userRepo.List(ctx, repositories.Pagination{
		Page:    page,
		PerPage: perPage,
	})
}
