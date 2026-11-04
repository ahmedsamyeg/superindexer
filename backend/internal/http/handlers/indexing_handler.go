package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// IndexingJobHandler handles indexing job management endpoints.
type IndexingJobHandler struct {
	DB   *bun.DB
	Log  interface{}
	Repo repositories.IndexingJobRepository
}

// CreateIndexingJobRequest represents a request to create an indexing job.
type CreateIndexingJobRequest struct {
	JobType     string  `json:"job_type" validate:"required"`
	SourceID    *int64  `json:"source_id"`
	NewsgroupID *int64  `json:"newsgroup_id"`
	Priority    int     `json:"priority"`
}

// List handles listing indexing jobs.
func (h *IndexingJobHandler) List(c echo.Context) error {
	status := c.QueryParam("status")
	var jobs []*models.IndexingJob
	var err error

	if status != "" {
		jobs, err = h.Repo.GetByStatus(c.Request().Context(), status)
	} else {
		jobs, err = h.Repo.GetPendingJobs(c.Request().Context())
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list jobs: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  jobs,
		"total": int64(len(jobs)),
	})
}

// GetByID handles getting an indexing job by ID.
func (h *IndexingJobHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job ID")
	}

	job, err := h.Repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"job": job,
	})
}

// Create handles creating a new indexing job.
func (h *IndexingJobHandler) Create(c echo.Context) error {
	var req CreateIndexingJobRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	job := &models.IndexingJob{
		JobType:     req.JobType,
		SourceID:    req.SourceID,
		NewsgroupID: req.NewsgroupID,
		Status:      "pending",
		Priority:    req.Priority,
		CreatedAt:   time.Now().UTC(),
	}

	if err := h.Repo.Create(c.Request().Context(), job); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to create job: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"job": job,
	})
}

// Retry handles retrying a failed indexing job.
func (h *IndexingJobHandler) Retry(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job ID")
	}

	if err := h.Repo.UpdateStatus(c.Request().Context(), id, "pending"); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "job retry initiated",
		"job_id":  id,
	})
}

// Cancel handles cancelling an indexing job.
func (h *IndexingJobHandler) Cancel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job ID")
	}

	if err := h.Repo.UpdateStatus(c.Request().Context(), id, "cancelled"); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "job cancelled",
		"job_id":  id,
	})
}