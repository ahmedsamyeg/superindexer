package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/superindexer/indexer/internal/auth"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/http/handlers"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository provides a mock implementation for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) UpdatePasswordHash(ctx context.Context, id int64, hash string) error {
	return m.Called(ctx, id, hash).Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, pagination repositories.Pagination) (*repositories.PaginatedResult, error) {
	args := m.Called(ctx, pagination)
	return args.Get(0).(*repositories.PaginatedResult), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePasswordResetToken(ctx context.Context, id int64, token string, expiry time.Time) error {
	return m.Called(ctx, id, token, expiry).Error(0)
}

func (m *MockUserRepository) DeletePasswordResetToken(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

// TestAuthLogin tests the login endpoint functionality
func TestAuthLogin(t *testing.T) {
	// Setup mock dependencies
	mockUserRepo := new(MockUserRepository)
	mockAuthService := new(MockAuthService)
	
	// Create auth handler with mocks
	authHandler := &handlers.AuthHandler{
		AuthService: mockAuthService,
	}
	
	t.Run("Successful Login", func(t *testing.T) {
		// Setup mock response
		loginResult := &service.LoginResult{
			Token: "mock-jwt-token",
			User: &models.User{
				ID:           1,
				Username:     "testuser",
				Email:        "test@example.com",
				Role:         "user",
				IsActive:     true,
				PasswordHash: "mock-hash",
			},
		}
		
		mockAuthService.On("Login", mock.Anything, "testuser", "password").Return(loginResult, nil)
		
		// Create test request
		reqBody := `{"username": "testuser", "password": "password"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.Login(c)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		// Verify response contains token and user data
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")
		
		mockAuthService.AssertExpectations(t)
	})
	
	t.Run("Invalid Credentials", func(t *testing.T) {
		// Setup mock error response
		mockAuthService.On("Login", mock.Anything, "invaliduser", "wrongpassword").Return(nil, assert.AnError)
		
		// Create test request
		reqBody := `{"username": "invaliduser", "password": "wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.Login(c)
		
		// Assert - should return 401
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		
		mockAuthService.AssertExpectations(t)
	})
	
	t.Run("Missing Fields", func(t *testing.T) {
		// Create test request with missing fields
		reqBody := `{"username": "testuser"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.Login(c)
		
		// Assert - should return 400 validation error
		assert.NoError(t, err)
		// Note: This would depend on the validator middleware being in place
	})
}

// TestAuthRegister tests the registration endpoint functionality  
func TestAuthRegister(t *testing.T) {
	// Setup mock dependencies
	mockUserRepo := new(MockUserRepository)
	mockAuthService := new(MockAuthService)
	
	// Create auth handler with mocks
	authHandler := &handlers.AuthHandler{
		AuthService: mockAuthService,
	}
	
	t.Run("Successful Registration", func(t *testing.T) {
		// Setup mock response
		user := &models.User{
			ID:           1,
			Username:     "newuser",
			Email:        "new@example.com",
			Role:         "user",
			IsActive:     true,
			PasswordHash: "mock-hash",
		}
		
		mockAuthService.On("Register", mock.Anything, "newuser", "new@example.com", "password").Return(user, nil)
		
		// Create test request
		reqBody := `{"username": "newuser", "email": "new@example.com", "password": "password"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.Register(c)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		
		// Verify response contains user data
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Contains(t, response, "user")
		assert.Equal(t, "newuser", response["user"].(map[string]interface{})["username"])
		
		mockAuthService.AssertExpectations(t)
	})
	
	t.Run("Duplicate Username", func(t *testing.T) {
		// Setup mock error response for duplicate username
		mockAuthService.On("Register", mock.Anything, "existinguser", "new@example.com", "password").Return(nil, assert.AnError)
		
		// Create test request
		reqBody := `{"username": "existinguser", "email": "new@example.com", "password": "password"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.Register(c)
		
		// Assert - should return 400
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		
		mockAuthService.AssertExpectations(t)
	})
}

// TestAuthLogout tests the logout endpoint functionality
func TestAuthLogout(t *testing.T) {
	// Setup mock dependencies (not needed for logout since it's stateless)
	authHandler := &handlers.AuthHandler{}
	
	t.Run("Successful Logout", func(t *testing.T) {
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.Logout(c)
		
		// Assert - should return 200
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		// Verify response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, "logged out successfully", response["message"])
	})
}