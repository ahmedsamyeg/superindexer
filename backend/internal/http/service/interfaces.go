package service

import (
	"context"

	"github.com/superindexer/indexer/internal/auth"
	"github.com/superindexer/indexer/internal/database/models"
)

// AuthServiceInterface defines the methods for authentication operations.
type AuthServiceInterface interface {
	Login(ctx context.Context, username, password string) (*LoginResult, error)
	Register(ctx context.Context, username, email, password string) (*models.User, error)
	ValidateToken(ctx context.Context, tokenString string) (*auth.TokenClaims, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}

// Ensure AuthService implements AuthServiceInterface.
var _ AuthServiceInterface = (*AuthService)(nil)

// UserServiceInterface defines the methods for user management operations.
type UserServiceInterface interface {
	CreateUser(ctx context.Context, username, email, password, role string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	ListUsers(ctx context.Context, page, perPage int) (*PaginatedResult, error)
	UpdateUser(ctx context.Context, id int64, username, email, password, role *string) (*models.User, error)
	DeleteUser(ctx context.Context, id int64) error
	ToggleUserActive(ctx context.Context, id int64) (*models.User, error)
}

// PaginatedResult holds paginated query results.
type PaginatedResult = models.PaginatedResult

// Ensure UserService implements UserServiceInterface.
var _ UserServiceInterface = (*UserService)(nil)
