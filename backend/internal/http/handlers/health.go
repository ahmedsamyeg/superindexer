package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthCheckResponse represents the health check response.
type HealthCheckResponse struct {
	Healthy   bool      `json:"healthy"`
	Timestamp time.Time `json:"timestamp"`
	Checks    []Check   `json:"checks"`
}

// Check represents a single health check.
type Check struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "healthy", "warning", "critical"
	Message string `json:"message"`
}

// HealthCheck returns the system health status.
func (h *Handlers) HealthCheck(c echo.Context) error {
	// Check database connectivity
	dbStatus := "healthy"
	dbMessage := "Database connection established"
	
	err := h.DB.Ping()
	if err != nil {
		dbStatus = "critical"
		dbMessage = "Database connection failed: " + err.Error()
	}
	
	healthy := dbStatus == "healthy"
	
	checks := []Check{
		{
			Name:    "Database",
			Status:  dbStatus,
			Message: dbMessage,
		},
		{
			Name:    "API",
			Status:  "healthy",
			Message: "API service is running",
		},
	}
	
	if dbStatus == "critical" {
		healthy = false
	}
	
	response := HealthCheckResponse{
		Healthy:   healthy,
		Timestamp: time.Now(),
		Checks:    checks,
	}
	
	statusCode := http.StatusOK
	if !healthy {
		statusCode = http.StatusServiceUnavailable
	}
	
	return c.JSON(statusCode, response)
}

// ReadyCheck checks if the system is ready to serve requests.
func (h *Handlers) ReadyCheck(c echo.Context) error {
	// Check database connectivity
	dbReady := false
	if err := h.DB.Ping(); err == nil {
		dbReady = true
	}
	
	ready := dbReady
	
	var messages []string
	if dbReady {
		messages = append(messages, "Database is connected")
	} else {
		messages = append(messages, "Database is not connected")
	}
	messages = append(messages, "API service is running")
	
	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}
	
	return c.JSON(statusCode, map[string]interface{}{
		"ready":   ready,
		"message": "System is ready to serve requests",
		"checks": map[string]bool{
			"database": dbReady,
		},
		"details": messages,
	})
}

// Version returns the application version.
func (h *Handlers) Version(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"version": "1.0.0",
		"build":   "2024-11-01",
	})
}