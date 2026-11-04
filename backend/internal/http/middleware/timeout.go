package middleware

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

// TimeoutConfig holds the configuration for the Timeout middleware.
type TimeoutConfig struct {
	Timeout time.Duration
}

// TimeoutWithConfig returns a timeout middleware.
func TimeoutWithConfig(cfg TimeoutConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), cfg.Timeout)
			defer cancel()

			req := c.Request().WithContext(ctx)
			c.SetRequest(req)

			errChan := make(chan error, 1)
			go func() {
				errChan <- next(c)
			}()

			select {
			case err := <-errChan:
				return err
			case <-ctx.Done():
				return echo.NewHTTPError(504, "request timed out").SetInternal(ctx.Err())
			}
		}
	}
}

// Recover returns a recovery middleware.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				err := recover()
				if err != nil {
					// Log the panic
					c.Logger().Error("panic recovered", "error", err)
					// Return a 500 error
					if !c.Response().Committed {
						_ = c.JSON(500, map[string]interface{}{
							"error": "internal server error",
						})
					}
				}
			}()
			return next(c)
		}
	}
}

// LoggerConfig holds the configuration for the Logger middleware.
type LoggerConfig struct {
	Skipper func(c echo.Context) bool
}

// LoggerWithConfig returns a structured logging middleware.
func LoggerWithConfig(cfg LoggerConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			// Skip logging for health checks
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return err
			}

			// Extract relevant fields
			method := c.Request().Method
			path := c.Request().URL.Path
			query := c.Request().URL.RawQuery
			status := c.Response().Status
			latency := time.Since(start)
			ip := c.RealIP()
			userAgent := c.Request().UserAgent()
			requestID := GetRequestID(c.Request().Context())

			// Log the request using slog
		entry := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
			With("method", method).
			With("path", path).
			With("query", query).
			With("status", status).
			With("latency", latency.String()).
			With("ip", ip).
			With("user_agent", userAgent).
			With("request_id", requestID)

			if err != nil {
				entry.Error("request error", "error", err.Error())
			} else {
				entry.Info("request completed")
			}

			return err
		}
	}
}

// CORSConfig holds the configuration for the CORS middleware.
type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// CORSWithConfig returns a CORS middleware.
func CORSWithConfig(cfg CORSConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Handle preflight requests
			if c.Request().Method == "OPTIONS" {
				for _, origin := range cfg.AllowOrigins {
					c.Response().Header().Set("Access-Control-Allow-Origin", origin)
				}
				c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
				c.Response().Header().Set("Access-Control-Max-Age", "86400")
				return c.NoContent(204)
			}

			// Set CORS headers for regular requests
			if len(cfg.AllowOrigins) > 0 {
				origin := c.Request().Header.Get("Origin")
				for _, allowed := range cfg.AllowOrigins {
					if allowed == "*" || allowed == origin {
						c.Response().Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

			return next(c)
		}
	}
}

// ValidateError checks if the error is a HTTP error.
func ValidateError(err error) bool {
	if err == nil {
		return false
	}
	var he *echo.HTTPError
	return errors.As(err, &he)
}
