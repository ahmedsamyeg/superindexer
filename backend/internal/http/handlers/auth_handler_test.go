package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/auth"
	"github.com/superindexer/indexer/internal/config"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uptrace/bun"
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

// MockAuthService provides a mock implementation for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*service.LoginResult, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*service.LoginResult), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, username, email, password string) (*models.User, error) {
	args := m.Called(ctx, username, email, password)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, tokenString string) (*auth.TokenClaims, error) {
	args := m.Called(ctx, tokenString)
	return args.Get(0).(*auth.TokenClaims), args.Error(1)
}

func (m *MockAuthService) RequestPasswordReset(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

func (m *MockAuthService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	return m.Called(ctx, token, newPassword).Error(0)
}

// TestAuthHandler tests authentication handler functionality
func TestAuthHandler(t *testing.T) {
	// Setup
	e := echo.New()
	
	// Create mock repositories and services
	mockUserRepo := new(MockUserRepository)
	mockAuthService := new(MockAuthService)
	
	// Create auth handler with mocks
	authHandler := &AuthHandler{
		AuthService: mockAuthService,
	}
	
	t.Run("Login Success", func(t *testing.T) {
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
		
		mockAuthService.AssertExpectations(t)
	})
	
	t.Run("Register Success", func(t *testing.T) {
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
		
		mockAuthService.AssertExpectations(t)
	})
	
	t.Run("RequestPasswordReset Success", func(t *testing.T) {
		// Setup mock response
		mockAuthService.On("RequestPasswordReset", mock.Anything, "test@example.com").Return(nil)
		
		// Create test request
		reqBody := `{"email": "test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.RequestPasswordReset(c)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		mockAuthService.AssertExpectations(t)
	})
	
	t.Run("ConfirmPasswordReset Success", func(t *testing.T) {
		// Setup mock response
		mockAuthService.On("ConfirmPasswordReset", mock.Anything, "mock-reset-token", "newpassword").Return(nil)
		
		// Create test request
		reqBody := `{"token": "mock-reset-token", "password": "newpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/confirm", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Execute
		err := authHandler.ConfirmPasswordReset(c)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		
		mockAuthService.AssertExpectations(t)
	})
	
	t.Run("GetMe Success", func(t *testing.T) {
		// Create test request with valid JWT token
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		// Add context with user info (this would normally be done by middleware)
		ctx := context.WithValue(c.Request().Context(), "user_id", int64(1))
		ctx = context.WithValue(ctx, "role", "user")
		ctx = context.WithValue(ctx, "username", "testuser")
		c.SetRequest(c.Request().WithContext(ctx))
		
		// Execute
		err := authHandler.GetMe(c)
		
		// Assert - should return success since we're just checking the structure
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// TestJWTValidation tests JWT token validation middleware
func TestJWTValidation(t *testing.T) {
	t.Run("Valid JWT Token", func(t *testing.T) {
		// Create a valid JWT token for testing
		claims := auth.TokenClaims{
			UserID:   1,
			Role:     "user",
			Username: "testuser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "superindexer",
			},
		}
		
		// Create JWT config with test secret
		jwtConfig := auth.JWTConfig{
			Secret: []byte("test-secret-key"),
			Issuer: "superindexer",
			Expiry: 3600,
		}
		
		token, err := auth.GenerateToken(1, "user", "testuser", jwtConfig)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		
		// Validate the token
		validatedClaims, err := auth.ValidateToken(token, jwtConfig)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), validatedClaims.UserID)
		assert.Equal(t, "user", validatedClaims.Role)
		assert.Equal(t, "testuser", validatedClaims.Username)
	})
}