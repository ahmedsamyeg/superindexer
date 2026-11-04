package middleware

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/auth"
	"github.com/uptrace/bun"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	userRoleKey  contextKey = "user_role"
	ContextKeyUsername contextKey = "username"
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
// It accepts a bun.DB for potential token revocation checks and a JWT secret.
func AuthMiddleware(db *bun.DB, jwtSecret string) echo.MiddlewareFunc {
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

			tokenString := parts[1]

			// Validate JWT token and extract user info
			claims, err := auth.ValidateToken(tokenString, auth.JWTConfig{
				Secret: []byte(jwtSecret),
				Issuer: "superindexer",
				Expiry: 3600, // 1 hour
			})
			if err != nil {
				return echo.NewHTTPError(401, "invalid token: "+err.Error())
			}

			// Store user info in context
			ctx := context.WithValue(c.Request().Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userRoleKey, claims.Role)
			ctx = context.WithValue(ctx, ContextKeyUsername, claims.Username)
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
func GetUserID(ctx context.Context) (int64, bool) {
	if id, ok := ctx.Value(userIDKey).(int64); ok {
		return id, true
	}
	return 0, false
}

// GetUserRole extracts the user role from context.
func GetUserRole(ctx context.Context) (string, bool) {
	if role, ok := ctx.Value(userRoleKey).(string); ok {
		return role, true
	}
	return "", false
}

// GetUsername extracts the username from context.
func GetUsername(ctx context.Context) (string, bool) {
	if username, ok := ctx.Value(ContextKeyUsername).(string); ok {
		return username, true
	}
	return "", false
}

// SetUserInContext adds user info to the request context (for testing).
func SetUserInContext(ctx context.Context, userID int64, role, username string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, userRoleKey, role)
	ctx = context.WithValue(ctx, ContextKeyUsername, username)
	return ctx
}