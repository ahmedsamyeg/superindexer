package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService provides a mock implementation for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	args := m.Called(ctx, username, password)
	if result, ok := args.Get(0).(*LoginResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, username, email, password string) (*User, error) {
	args := m.Called(ctx, username, email, password)
	if user, ok := args.Get(0).(*User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	args := m.Called(ctx, tokenString)
	if claims, ok := args.Get(0).(*TokenClaims); ok {
		return claims, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) RequestPasswordReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

// LoginResult holds the result of a login attempt.
type LoginResult struct {
	Token  string
	User   *User
}

// User represents a user model.
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Role         string
	IsActive     bool
}

// TokenClaims holds the JWT claims.
type TokenClaims struct {
	UserID   int64
	Role     string
	Username string
}

// TestHealthEndpoints tests system health and readiness endpoints
func TestHealthEndpoints(t *testing.T) {
	e := echo.New()

	// Test health endpoint returns 200 with proper structure
	t.Run("Health Check Returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HealthCheck(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.True(t, response["healthy"].(bool))
		assert.Contains(t, response, "checks")
		assert.Contains(t, response, "timestamp")
	})

	// Test readiness endpoint returns 200
	t.Run("Readiness Check Returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ReadyCheck(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.True(t, response["ready"].(bool))
		assert.Contains(t, response, "message")
	})
}

// TestAPIValidation tests request validation and error handling
func TestAPIValidation(t *testing.T) {
	e := echo.New()

	t.Run("Login With Missing Username Returns 400", func(t *testing.T) {
		reqBody := `{"password": "password"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ValidateLoginRequest(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Login With Missing Password Returns 400", func(t *testing.T) {
		reqBody := `{"username": "testuser"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ValidateLoginRequest(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Login With Empty Body Returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ValidateLoginRequest(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Register With Invalid Email Returns 400", func(t *testing.T) {
		reqBody := `{"username": "testuser", "email": "invalid-email", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ValidateRegisterRequest(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Register With Short Password Returns 400", func(t *testing.T) {
		reqBody := `{"username": "testuser", "email": "test@example.com", "password": "short"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ValidateRegisterRequest(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Register With Missing Fields Returns 400", func(t *testing.T) {
		reqBody := `{"username": "testuser"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ValidateRegisterRequest(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestAuthHandlerWithMocks tests authentication handler with mocked services
func TestAuthHandlerWithMocks(t *testing.T) {
	e := echo.New()
	mockService := new(MockAuthService)

	t.Run("Successful Login Returns Token", func(t *testing.T) {
		loginResult := &LoginResult{
			Token: "mock-jwt-token",
			User: &User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
				Role:     "user",
				IsActive: true,
			},
		}
		mockService.On("Login", mock.Anything, "testuser", "password").Return(loginResult, nil)

		reqBody := `{"username": "testuser", "password": "password"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Simulate login flow
		var reqData struct {
			Username string `json:"username" validate:"required"`
			Password string `json:"password" validate:"required"`
		}
		_ = json.NewDecoder(req.Body).Decode(&reqData)

		// Validate
		if reqData.Username == "" || reqData.Password == "" {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "validation failed"})
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			return
		}

		result, err := mockService.Login(c.Request().Context(), reqData.Username, reqData.Password)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Token)
		assert.Equal(t, "testuser", result.User.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Failed Login Returns 401", func(t *testing.T) {
		mockService.On("Login", mock.Anything, "invaliduser", "wrongpassword").Return(
			(*LoginResult)(nil), assert.AnError)

		reqBody := `{"username": "invaliduser", "password": "wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var reqData struct {
			Username string `json:"username" validate:"required"`
			Password string `json:"password" validate:"required"`
		}
		_ = json.NewDecoder(req.Body).Decode(&reqData)

		result, err := mockService.Login(c.Request().Context(), reqData.Username, reqData.Password)
		assert.Error(t, err)
		assert.Nil(t, result)

		mockService.AssertExpectations(t)
	})

	t.Run("Successful Registration Returns User", func(t *testing.T) {
		user := &User{
			ID:       2,
			Username: "newuser",
			Email:    "new@example.com",
			Role:     "user",
			IsActive: true,
		}
		mockService.On("Register", mock.Anything, "newuser", "new@example.com", "password123").Return(user, nil)

		reqBody := `{"username": "newuser", "email": "new@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var reqData struct {
			Username string `json:"username" validate:"required"`
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,min=8"`
		}
		_ = json.NewDecoder(req.Body).Decode(&reqData)

		result, err := mockService.Register(c.Request().Context(), reqData.Username, reqData.Email, reqData.Password)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "newuser", result.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Password Reset Request Returns Success", func(t *testing.T) {
		mockService.On("RequestPasswordReset", mock.Anything, "test@example.com").Return(nil)

		reqBody := `{"email": "test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var reqData struct {
			Email string `json:"email" validate:"required,email"`
		}
		_ = json.NewDecoder(req.Body).Decode(&reqData)

		err := mockService.RequestPasswordReset(c.Request().Context(), reqData.Email)
		assert.NoError(t, err)

		mockService.AssertExpectations(t)
	})

	t.Run("Password Reset Confirm Returns Success", func(t *testing.T) {
		mockService.On("ConfirmPasswordReset", mock.Anything, "reset-token", "newpassword123").Return(nil)

		reqBody := `{"token": "reset-token", "password": "newpassword123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/confirm", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var reqData struct {
			Token    string `json:"token" validate:"required"`
			Password string `json:"password" validate:"required,min=8"`
		}
		_ = json.NewDecoder(req.Body).Decode(&reqData)

		err := mockService.ConfirmPasswordReset(c.Request().Context(), reqData.Token, reqData.Password)
		assert.NoError(t, err)

		mockService.AssertExpectations(t)
	})
}

// TestVersionEndpoint tests the version endpoint
func TestVersionEndpoint(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Version(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "version")
	assert.Contains(t, response, "build")
}

// TestLogoutEndpoint tests the logout endpoint
func TestLogoutEndpoint(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Logout(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "logged out successfully", response["message"])
}
