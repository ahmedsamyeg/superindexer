package coordinator

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/indexing/workers"
	"log/slog"
)

// IndexingCoordinator manages the overall indexing process
type IndexingCoordinator struct {
	jobRepo       repositories.IndexingJobRepository
	sourceRepo    repositories.SpotnetSourceRepository
	newsgroupRepo repositories.NewsgroupRepository
	workerPool    *workers.WorkerPool
	jobQueue      chan *models.IndexingJob
	pendingQueue  []*models.IndexingJob
	mu            sync.RWMutex
	isRunning     bool
	stopCh        chan struct{}
	log           *slog.Logger
}

// NewIndexingCoordinator creates a new indexing coordinator
func NewIndexingCoordinator(
	jobRepo repositories.IndexingJobRepository,
	sourceRepo repositories.SpotnetSourceRepository,
	newsgroupRepo repositories.NewsgroupRepository,
	workerPool *workers.WorkerPool,
) *IndexingCoordinator {
	return &IndexingCoordinator{
		jobRepo:       jobRepo,
		sourceRepo:    sourceRepo,
		newsgroupRepo: newsgroupRepo,
		workerPool:    workerPool,
		jobQueue:      make(chan *models.IndexingJob, 100),
		pendingQueue:  make([]*models.IndexingJob, 0),
		stopCh:        make(chan struct{}),
		log:           slog.Default(),
	}
}

// Start initializes and starts the indexing coordinator
func (c *IndexingCoordinator) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		c.log.Info("IndexingCoordinator already running")
		return nil
	}

	c.isRunning = true
	c.stopCh = make(chan struct{})

	c.log.Info("Starting indexing coordinator")

	// Start the background job processing goroutine
	go c.processJobs(ctx)

	// Start the worker pool
	if err := c.workerPool.Start(ctx, nil); err != nil {
		c.isRunning = false
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	c.log.Info("Indexing coordinator started successfully")
	return nil
}

// Stop shuts down the indexing coordinator gracefully
func (c *IndexingCoordinator) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		c.log.Info("IndexingCoordinator already stopped")
		return nil
	}

	c.log.Info("Stopping indexing coordinator")

	// Signal shutdown
	close(c.stopCh)

	c.isRunning = false

	// Stop the worker pool
	if err := c.workerPool.Stop(ctx); err != nil {
		c.log.Error("Failed to stop worker pool", "error", err)
	}

	// Clear pending queue
	c.mu.Lock()
	c.pendingQueue = c.pendingQueue[:0]
	c.mu.Unlock()

	c.log.Info("Indexing coordinator stopped")
	return nil
}

// SubmitJob creates and submits a new indexing job to the coordinator
func (c *IndexingCoordinator) SubmitJob(ctx context.Context, jobType string, sourceID, newsgroupID *int64) (*models.IndexingJob, error) {
	// Validate job type
	if jobType != "spotnet" && jobType != "nntp" {
		return nil, fmt.Errorf("invalid job type: %s", jobType)
	}

	// Create new job record
	job := &models.IndexingJob{
		JobType:     jobType,
		SourceID:    sourceID,
		NewsgroupID: newsgroupID,
		Status:      "pending",
		Priority:    c.calculatePriority(jobType, sourceID, newsgroupID),
		Progress:    0,
		RecordsProcessed: 0,
		RecordsFailed:    0,
		CreatedAt:      time.Now().UTC(),
	}

	// Persist job to database
	if err := c.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create indexing job: %w", err)
	}

	// Add to pending queue and sort by priority (higher first)
	c.mu.Lock()
	c.pendingQueue = append(c.pendingQueue, job)
	sort.Slice(c.pendingQueue, func(i, j int) bool {
		if c.pendingQueue[i].Priority != c.pendingQueue[j].Priority {
			return c.pendingQueue[i].Priority > c.pendingQueue[j].Priority
		}
		return c.pendingQueue[i].CreatedAt.Before(c.pendingQueue[j].CreatedAt)
	})
	c.mu.Unlock()

	c.log.Info("Submitted new indexing job",
		"job_id", job.ID,
		"job_type", jobType,
		"priority", job.Priority,
	)

	return job, nil
}

// GetJobStatus retrieves the status of a specific indexing job
func (c *IndexingCoordinator) GetJobStatus(ctx context.Context, jobID int64) (*models.IndexingJob, error) {
	job, err := c.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job %d: %w", jobID, err)
	}
	return job, nil
}

// ListJobs retrieves a list of indexing jobs with optional filtering
func (c *IndexingCoordinator) ListJobs(ctx context.Context, status string, page, perPage int) (*repositories.PaginatedResult, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Get all jobs (filtering happens at DB level via GetByStatus)
	var jobs []*models.IndexingJob
	var err error

	if status != "" {
		jobs, err = c.jobRepo.GetByStatus(ctx, status)
		if err != nil {
			return nil, fmt.Errorf("failed to get jobs with status %s: %w", status, err)
		}
	} else {
		// Get all statuses by querying each one and merging
		for _, s := range []string{"pending", "running", "completed", "failed", "cancelled"} {
			j, err := c.jobRepo.GetByStatus(ctx, s)
			if err != nil {
				continue
			}
			jobs = append(jobs, j...)
		}
	}

	// Sort by priority DESC, created_at ASC
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].Priority != jobs[j].Priority {
			return jobs[i].Priority > jobs[j].Priority
		}
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})

	total := int64(len(jobs))

	// Apply pagination
	start := (page - 1) * perPage
	if start >= len(jobs) {
		jobs = []*models.IndexingJob{}
	} else {
		end := start + perPage
		if end > len(jobs) {
			end = len(jobs)
		}
		jobs = jobs[start:end]
	}

	// Ensure non-nil slice for JSON marshaling
	if jobs == nil {
		jobs = []*models.IndexingJob{}
	}

	return &repositories.PaginatedResult{
		Data:       jobs,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}, nil
}

// CancelJob cancels a pending or running job
func (c *IndexingCoordinator) CancelJob(ctx context.Context, jobID int64) error {
	// Get the current job
	job, err := c.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job %d: %w", jobID, err)
	}

	// Check if job can be cancelled
	if job.Status == "completed" || job.Status == "cancelled" {
		return fmt.Errorf("cannot cancel job in %s status", job.Status)
	}

	// Remove from pending queue
	c.mu.Lock()
	for i, p := range c.pendingQueue {
		if p.ID == jobID {
			c.pendingQueue = append(c.pendingQueue[:i], c.pendingQueue[i+1:]...)
			break
		}
	}
	c.mu.Unlock()

	// Update job status in database
	if job.Status == "running" {
		now := time.Now().UTC()
		job.Status = "cancelled"
		job.CompletedAt = &now
		if err := c.jobRepo.Update(ctx, job); err != nil {
			return fmt.Errorf("failed to update job status to cancelled: %w", err)
		}
	} else {
		if err := c.jobRepo.UpdateStatus(ctx, jobID, "cancelled"); err != nil {
			return fmt.Errorf("failed to update job status to cancelled: %w", err)
		}
	}

	c.log.Info("Job cancelled", "job_id", jobID)
	return nil
}

// RetryJob retries a failed job
func (c *IndexingCoordinator) RetryJob(ctx context.Context, jobID int64) error {
	// Get the current job
	job, err := c.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job %d: %w", jobID, err)
	}

	// Only retry failed jobs
	if job.Status != "failed" {
		return fmt.Errorf("can only retry failed jobs, current status: %s", job.Status)
	}

	// Reset job state
	job.Status = "pending"
	job.Progress = 0
	job.RecordsProcessed = 0
	job.RecordsFailed = 0
	job.ErrorMessage = ""
	job.StartedAt = nil
	job.CompletedAt = nil

	if err := c.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to reset job state: %w", err)
	}

	// Add back to pending queue
	c.mu.Lock()
	c.pendingQueue = append(c.pendingQueue, job)
	sort.Slice(c.pendingQueue, func(i, j int) bool {
		if c.pendingQueue[i].Priority != c.pendingQueue[j].Priority {
			return c.pendingQueue[i].Priority > c.pendingQueue[j].Priority
		}
		return c.pendingQueue[i].CreatedAt.Before(c.pendingQueue[j].CreatedAt)
	})
	c.mu.Unlock()

	c.log.Info("Job resubmitted for retry", "job_id", jobID)
	return nil
}

// StartProcessing starts the worker pool for processing jobs
func (c *IndexingCoordinator) StartProcessing(ctx context.Context) error {
	if c.workerPool == nil {
		return fmt.Errorf("worker pool is not initialized")
	}

	if err := c.workerPool.Start(ctx, nil); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	c.log.Info("Worker pool started for processing")
	return nil
}

// StopProcessing stops the worker pool and waits for in-flight jobs to complete
func (c *IndexingCoordinator) StopProcessing() error {
	if c.workerPool == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.workerPool.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop worker pool: %w", err)
	}

	c.log.Info("Worker pool stopped")
	return nil
}

// processJobs is the background goroutine that processes jobs from the queue
func (c *IndexingCoordinator) processJobs(ctx context.Context) {
	c.log.Info("Job processor goroutine started")

	// Periodically check for pending jobs and submit them to the worker pool
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("Job processor shutting down", "reason", ctx.Err())
			return
		case <-c.stopCh:
			c.log.Info("Job processor received stop signal")
			return
		case <-ticker.C:
			c.processPendingJobs(ctx)
		}
	}
}

// processPendingJobs moves pending jobs from the priority queue to the worker pool
func (c *IndexingCoordinator) processPendingJobs(ctx context.Context) {
	c.mu.Lock()
	if len(c.pendingQueue) == 0 {
		c.mu.Unlock()
		return
	}

	// Take the highest priority job
	job := c.pendingQueue[0]
	c.pendingQueue = c.pendingQueue[1:]
	c.mu.Unlock()

	// Update job status to running
	now := time.Now().UTC()
	job.Status = "running"
	job.StartedAt = &now
	if err := c.jobRepo.Update(ctx, job); err != nil {
		c.log.Error("Failed to update job status to running", "job_id", job.ID, "error", err)
		return
	}

	c.log.Info("Processing job", "job_id", job.ID, "job_type", job.JobType)

	// Submit to worker pool (non-blocking)
	if err := c.workerPool.SubmitJob(ctx, job); err != nil {
		c.log.Error("Failed to submit job to worker pool", "job_id", job.ID, "error", err)

		// Mark as failed
		c.mu.Lock()
		c.pendingQueue = append(c.pendingQueue, job)
		c.mu.Unlock()

		if err := c.jobRepo.UpdateStatus(ctx, job.ID, "pending"); err != nil {
			c.log.Error("Failed to reset job status", "job_id", job.ID, "error", err)
		}
	}
}

// calculatePriority determines the priority of a job
func (c *IndexingCoordinator) calculatePriority(jobType string, sourceID, newsgroupID *int64) int {
	priority := 0

	// Base priority by job type
	switch jobType {
	case "spotnet":
		priority = 5
	case "nntp":
		priority = 3
	}

	// Boost priority based on source ID (lower ID = higher priority, older sources first)
	if sourceID != nil {
		priority += int(*sourceID % 10)
	}

	// Boost priority for newsgroups with lower indexing priority
	if newsgroupID != nil {
		// We'll use a heuristic: lower newsgroup ID = higher priority
		priority += int(*newsgroupID % 5)
	}

	return priority
}
