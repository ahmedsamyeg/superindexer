package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
	List(ctx context.Context, pagination Pagination) (*PaginatedResult, error)
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	// Password reset related methods
	GetByPasswordResetToken(ctx context.Context, token string) (*models.User, error)
	UpdatePasswordResetToken(ctx context.Context, id int64, token string, expiry time.Time) error
	DeletePasswordResetToken(ctx context.Context, id int64) error
}

// paginatedResult holds pagination parameters.
type Pagination struct {
	Page    int
	PerPage int
}

// PaginatedResult holds paginated query results.
type PaginatedResult struct {
	Data       interface{}
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

// newUserPaginatedResult creates a new paginated result.
func newUserPaginatedResult(data interface{}, total, page, perPage int64) *PaginatedResult {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}
	return &PaginatedResult{
		Data:       data,
		Total:      total,
		Page:       int(page),
		PerPage:    int(perPage),
		TotalPages: int(totalPages),
	}
}

// userRepository implements UserRepository using Bun.
type userRepository struct {
	db *bun.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user.
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

// GetByID retrieves a user by ID.
func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	err := r.db.NewSelect().Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByUsername retrieves a user by username.
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.NewSelect().Model(user).
		Where("username = ?", username).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.NewSelect().Model(user).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates an existing user.
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.NewUpdate().Model(user).Where("id = ?", user.ID).Exec(ctx)
	return err
}

// UpdatePasswordHash updates only the password hash.
func (r *userRepository) UpdatePasswordHash(ctx context.Context, id int64, hash string) error {
	_, err := r.db.NewUpdate().
		Model(&models.User{}).
		Set("password_hash = ?", hash).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// GetByPasswordResetToken retrieves a user by password reset token.
func (r *userRepository) GetByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	user := &models.User{}
	err := r.db.NewSelect().Model(user).
		Where("password_reset_token = ?", token).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdatePasswordResetToken updates the password reset token and its expiry for a user.
func (r *userRepository) UpdatePasswordResetToken(ctx context.Context, id int64, token string, expiry time.Time) error {
	_, err := r.db.NewUpdate().
		Model(&models.User{}).
		Set("password_reset_token = ?, password_reset_expiry = ?", token, expiry).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// DeletePasswordResetToken removes the password reset token from a user.
func (r *userRepository) DeletePasswordResetToken(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		Model(&models.User{}).
		Set("password_reset_token = ?, password_reset_expiry = ?", nil, nil).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// List retrieves a paginated list of users.
func (r *userRepository) List(ctx context.Context, pagination Pagination) (*PaginatedResult, error) {
	var users []models.User
	query := r.db.NewSelect().Model(&users)

	// Count total
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	query = query.
		Offset(int64((pagination.Page - 1) * pagination.PerPage)).
		Limit(int64(pagination.PerPage)).
		Order("created_at DESC")

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return newUserPaginatedResult(users, total, int64(pagination.Page), int64(pagination.PerPage)), nil
}

// Delete removes a user by ID.
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().
		Model(&models.User{}).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("user not found: %d", id)
	}
	return nil
}

// Count returns the total number of users.
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.db.NewSelect().Model(&models.User{}).Count(ctx)
}

// Ensure interface compliance
var _ UserRepository = (*userRepository)(nil)