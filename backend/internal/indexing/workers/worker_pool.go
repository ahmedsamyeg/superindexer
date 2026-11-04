package workers

import (
	"context"
	"fmt"
	"sync"

	"github.com/superindexer/indexer/internal/database/models"
)

// WorkerPool manages a pool of indexing workers
type WorkerPool struct {
	workers    []*Worker
	jobQueue   chan *models.IndexingJob
	wg         sync.WaitGroup
	isRunning  bool
	mu         sync.RWMutex
	maxWorkers int
}

// NewWorkerPool creates a new worker pool with specified maximum workers
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		workers:    make([]*Worker, 0, maxWorkers),
		jobQueue:   make(chan *models.IndexingJob, 100), // Buffered channel for jobs
		maxWorkers: maxWorkers,
		isRunning:  false,
	}
}

// Start initializes and starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context, workers []*Worker) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.isRunning {
		return nil // Already running
	}

	wp.workers = workers
	wp.isRunning = true

	// Start worker goroutines to process jobs from the queue
	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go func(w *Worker) {
			defer wp.wg.Done()
			w.Run(ctx, wp.jobQueue)
		}(worker)
	}

	return nil
}

// Stop shuts down the worker pool gracefully
func (wp *WorkerPool) Stop(ctx context.Context) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.isRunning {
		return nil // Already stopped
	}

	wp.isRunning = false
	
	// Close the job queue to signal workers to stop
	close(wp.jobQueue)

	// Wait for all workers to finish processing their current jobs
	wp.wg.Wait()

	return nil
}

// SubmitJob submits a job to the worker pool for processing
func (wp *WorkerPool) SubmitJob(ctx context.Context, job *models.IndexingJob) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if !wp.isRunning {
		return fmt.Errorf("worker pool is not running")
	}

	// Non-blocking send to job queue
	select {
	case wp.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}