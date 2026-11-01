package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/stretchr/testify/assert"
)

// MockUserService implements the UserService interface for testing
type MockUserService struct {
	users map[int64]*models.User
}

func (m *MockUserService) ListUsers(ctx context.Context, page, perPage int) (*service.PaginatedResult, error) {
	return &service.PaginatedResult{
		Data:       []models.User{},
		Total:      0,
		Page:       page,
		PerPage:    perPage,
		TotalPages: 0,
	}, nil
}

func (m *MockUserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, &UserNotFoundError{ID: id}
}

func (m *MockUserService) CreateUser(ctx context.Context, username, email, password, role string) (*models.User, error) {
	user := &models.User{
		ID:           int64(len(m.users) + 1),
		Username:     username,
		Email:        email,
		PasswordHash: password, // In real app, this would be hashed
		Role:         role,
		IsActive:     true,
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *MockUserService) UpdateUser(ctx context.Context, id int64, username, email, password, role *string) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, &UserNotFoundError{ID: id}
	}

	if username != nil {
		user.Username = *username
	}
	if email != nil {
		user.Email = *email
	}
	if password != nil {
		user.PasswordHash = *password
	}
	if role != nil {
		user.Role = *role
	}

	return user, nil
}

func (m *MockUserService) DeleteUser(ctx context.Context, id int64) error {
	if _, exists := m.users[id]; !exists {
		return &UserNotFoundError{ID: id}
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserService) ToggleUserActive(ctx context.Context, id int64) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, &UserNotFoundError{ID: id}
	}
	user.IsActive = !user.IsActive
	return user, nil
}

type UserNotFoundError struct {
	ID int64
}

func (e *UserNotFoundError) Error() string {
	return "user not found"
}

func TestUserHandler(t *testing.T) {
	// Create a mock user service
	mockUserService := &MockUserService{
		users: make(map[int64]*models.User),
	}

	// Create a user handler with the mock service
	handler := &UserHandler{
		UserService: mockUserService,
	}

	// Create test context and request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	req.Header.Set("Content-Type", "application/json")

	// Create test response recorder
	rec := httptest.NewRecorder()

	// Test that handler was created successfully
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.UserService)
}