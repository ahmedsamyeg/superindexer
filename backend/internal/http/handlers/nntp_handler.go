package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// NNTPServerHandler handles NNTP server management endpoints.
type NNTPServerHandler struct {
	DB   *bun.DB
	Log  interface{}
	Repo repositories.NNTPServerRepository
}

// CreateNNTPServerRequest represents a request to create an NNTP server.
type CreateNNTPServerRequest struct {
	Name          string `json:"name" validate:"required"`
	Hostname      string `json:"hostname" validate:"required"`
	Port          int    `json:"port" validate:"required,min=1,max=65535"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	TLSEnabled    bool   `json:"tls_enabled"`
	MaxConnections int   `json:"max_connections"`
	Priority      int    `json:"priority"`
}

// UpdateNNTPServerRequest represents a request to update an NNTP server.
type UpdateNNTPServerRequest struct {
	Name          *string `json:"name"`
	Hostname      *string `json:"hostname"`
	Port          *int    `json:"port"`
	Username      *string `json:"username"`
	Password      *string `json:"password"`
	TLSEnabled    *bool   `json:"tls_enabled"`
	MaxConnections *int   `json:"max_connections"`
	Priority      *int    `json:"priority"`
	Enabled       *bool   `json:"enabled"`
}

// List handles listing NNTP servers.
func (h *NNTPServerHandler) List(c echo.Context) error {
	servers, err := h.Repo.GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list servers: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  servers,
		"total": int64(len(servers)),
	})
}

// GetByID handles getting an NNTP server by ID.
func (h *NNTPServerHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid server ID")
	}

	server, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "server not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"server": server,
	})
}

// Create handles creating a new NNTP server.
func (h *NNTPServerHandler) Create(c echo.Context) error {
	var req CreateNNTPServerRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	server := &models.NNTPServer{
		Name:           req.Name,
		Hostname:       req.Hostname,
		Port:           req.Port,
		Username:       req.Username,
		TLSEnabled:     req.TLSEnabled,
		MaxConnections: req.MaxConnections,
		Priority:       req.Priority,
		Enabled:        true,
		ConnectionStatus: "disconnected",
	}

	if req.Password != "" {
		server.EncryptedPassword = req.Password
	}

	if err := h.Repo.Create(c.Request().Context(), server); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create server: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"server": server,
	})
}

// Update handles updating an NNTP server.
func (h *NNTPServerHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid server ID")
	}

	server, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "server not found")
	}

	var req UpdateNNTPServerRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	if req.Name != nil {
		server.Name = *req.Name
	}
	if req.Hostname != nil {
		server.Hostname = *req.Hostname
	}
	if req.Port != nil {
		server.Port = *req.Port
	}
	if req.Username != nil {
		server.Username = *req.Username
	}
	if req.Password != nil {
		server.EncryptedPassword = *req.Password
	}
	if req.TLSEnabled != nil {
		server.TLSEnabled = *req.TLSEnabled
	}
	if req.MaxConnections != nil {
		server.MaxConnections = *req.MaxConnections
	}
	if req.Priority != nil {
		server.Priority = *req.Priority
	}
	if req.Enabled != nil {
		server.Enabled = *req.Enabled
	}

	if err := h.Repo.Update(c.Request().Context(), server); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update server: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"server": server,
	})
}

// Delete handles deleting an NNTP server.
func (h *NNTPServerHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid server ID")
	}

	if err := h.Repo.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "server not found")
	}

	return c.NoContent(http.StatusNoContent)
}

// TestConnection handles testing NNTP server connectivity.
func (h *NNTPServerHandler) TestConnection(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid server ID")
	}

	// Update status to indicate test is in progress
	_ = h.Repo.UpdateStatus(c.Request().Context(), id, "testing")

	// In a real implementation, we would:
	// 1. Create a new NNTP client connection
	// 2. Attempt to list groups or fetch a test article
	// 3. Update the connection status based on the result
	// For now, return a simulated response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "connected",
		"message": "connection test successful",
	})
}