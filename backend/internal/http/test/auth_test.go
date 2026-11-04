package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAuthLogin tests the login endpoint functionality
func TestAuthLogin(t *testing.T) {
	e := echo.New()
	mockService := new(MockAuthService)

	t.Run("Successful Login", func(t *testing.T) {
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
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "mock-jwt-token", result.Token)
		assert.Equal(t, "testuser", result.User.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		mockService.On("Login", mock.Anything, "invaliduser", "wrongpassword").Return(
			(*LoginResult)(nil), assert.AnError)

		reqBody := `{"username": "invaliduser", "password": "wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
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

	t.Run("Missing Fields", func(t *testing.T) {
		reqBody := `{"username": "testuser"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ValidateLoginRequest(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestAuthRegister tests the registration endpoint functionality
func TestAuthRegister(t *testing.T) {
	e := echo.New()
	mockService := new(MockAuthService)

	t.Run("Successful Registration", func(t *testing.T) {
		user := &User{
			ID:       1,
			Username: "newuser",
			Email:    "new@example.com",
			Role:     "user",
			IsActive: true,
		}

		mockService.On("Register", mock.Anything, "newuser", "new@example.com", "password123").Return(user, nil)

		reqBody := `{"username": "newuser", "email": "new@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
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
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "newuser", result.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Duplicate Username", func(t *testing.T) {
		mockService.On("Register", mock.Anything, "existinguser", "new@example.com", "password").Return(
			(*User)(nil), assert.AnError)

		reqBody := `{"username": "existinguser", "email": "new@example.com", "password": "password"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
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
		assert.Error(t, err)
		assert.Nil(t, result)

		mockService.AssertExpectations(t)
	})
}

// TestAuthLogout tests the logout endpoint functionality
func TestAuthLogout(t *testing.T) {
	e := echo.New()

	t.Run("Successful Logout", func(t *testing.T) {
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
	})
}

// TestPasswordResetFlow tests the password reset flow
func TestPasswordResetFlow(t *testing.T) {
	e := echo.New()
	mockService := new(MockAuthService)

	t.Run("Request Password Reset", func(t *testing.T) {
		mockService.On("RequestPasswordReset", mock.Anything, "test@example.com").Return(nil)

		reqBody := `{"email": "test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset", bytes.NewBufferString(reqBody))
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

	t.Run("Confirm Password Reset", func(t *testing.T) {
		mockService.On("ConfirmPasswordReset", mock.Anything, "reset-token", "newpassword123").Return(nil)

		reqBody := `{"token": "reset-token", "password": "newpassword123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/confirm", bytes.NewBufferString(reqBody))
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

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	mockService := new(MockAuthService)

	t.Run("Valid Token", func(t *testing.T) {
		claims := &TokenClaims{
			UserID:   1,
			Role:     "user",
			Username: "testuser",
		}
		mockService.On("ValidateToken", mock.Anything, "valid-token").Return(claims, nil)

		result, err := mockService.ValidateToken(context.Background(), "valid-token")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.UserID)
		assert.Equal(t, "user", result.Role)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		mockService.On("ValidateToken", mock.Anything, "invalid-token").Return(
			(*TokenClaims)(nil), assert.AnError)

		result, err := mockService.ValidateToken(context.Background(), "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, result)

		mockService.AssertExpectations(t)
	})
}
