package test

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var validate = validator.New()

// HealthCheck simulates the health check endpoint response.
func HealthCheck(c echo.Context) error {
	response := map[string]interface{}{
		"healthy":   true,
		"timestamp": "2024-01-01T00:00:00Z",
		"checks": []map[string]string{
			{"name": "Database", "status": "healthy", "message": "Database connection established"},
			{"name": "API", "status": "healthy", "message": "API service is running"},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// ReadyCheck simulates the readiness check endpoint response.
func ReadyCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"ready":   true,
		"message": "System is ready to serve requests",
	})
}

// Version returns the application version.
func Version(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"version": "1.0.0",
		"build":   "2024-11-01",
	})
}

// Logout handles user logout.
func Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "logged out successfully",
	})
}

// ValidateLoginRequest validates the login request body.
func ValidateLoginRequest(c echo.Context) error {
	var req struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if err := validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[strings.ToLower(err.Field())] = err.Tag()
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":             "validation failed",
			"validation_errors": validationErrors,
		})
	}

	return nil
}

// ValidateRegisterRequest validates the registration request body.
func ValidateRegisterRequest(c echo.Context) error {
	var req struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if err := validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[strings.ToLower(err.Field())] = err.Tag()
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":             "validation failed",
			"validation_errors": validationErrors,
		})
	}

	return nil
}
