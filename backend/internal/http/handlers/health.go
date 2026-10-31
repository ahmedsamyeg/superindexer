package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles the /health endpoint.
func (h *Handlers) HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
		"service": "indexer",
		"version": "0.1.0",
	})
}

// ReadyHandler handles the /ready endpoint.
func (h *Handlers) ReadyHandler(c echo.Context) error {
	// Check database connectivity
	if h.DB != nil {
		if err := h.DB.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status":  "not_ready",
				"reason":  "database unavailable",
				"error":   err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ready",
	})
}
