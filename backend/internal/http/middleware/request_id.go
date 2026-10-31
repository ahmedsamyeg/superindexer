package middleware

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	userRoleKey  contextKey = "user_role"
)

// RequestID generates a unique request ID and adds it to the context and response headers.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Store in context
			ctx := context.WithValue(c.Request().Context(), requestIDKey, requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Add to response header
			c.Response().Header().Set("X-Request-ID", requestID)

			return next(c)
		}
	}
}

// AuthMiddleware creates an Echo middleware that validates JWT tokens.
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(401, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(401, "invalid authorization header format")
			}

			_ = parts[1]

			// TODO: Validate JWT token and extract user info
			// For now, set a placeholder user ID
			ctx := context.WithValue(c.Request().Context(), userIDKey, int64(1))
			ctx = context.WithValue(ctx, userRoleKey, "admin")
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// AdminMiddleware creates an Echo middleware that checks for admin role.
func AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Request().Context().Value(userRoleKey).(string)
			if !ok || role != "admin" {
				return echo.NewHTTPError(403, "admin access required")
			}
			return next(c)
		}
	}
}

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID extracts the user ID from context.
func GetUserID(ctx context.Context) int64 {
	if id, ok := ctx.Value(userIDKey).(int64); ok {
		return id
	}
	return 0
}

// GetUserRole extracts the user role from context.
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(userRoleKey).(string); ok {
		return role
	}
	return ""
}
