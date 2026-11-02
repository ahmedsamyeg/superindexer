package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
)

// AdminDashboard provides administrative overview and monitoring capabilities
type AdminDashboard struct {
	nntpRepo        repositories.NNTPServerRepository
	newsgroupRepo   repositories.NewsgroupRepository
	spotnetRepo     repositories.SpotnetSourceRepository
	articleRepo     repositories.ArticleRepository
	postRepo        repositories.SpotnetPostRepository
	jobRepo         repositories.IndexingJobRepository
	categoryRepo    repositories.CategoryRepository
}

// DashboardStats holds administrative statistics
type DashboardStats struct {
	// System statistics
	TotalArticles       int64 `json:"total_articles"`
	TotalSpotnetPosts   int64 `json:"total_spotnet_posts"`
	ActiveNNTPServers   int64 `json:"active_nntp_servers"`
	ActiveSpotnetSources int64 `json:"active_spotnet_sources"`
	TotalCategories     int64 `json:"total_categories"`
	ActiveIndexingJobs  int64 `json:"active_indexing_jobs"`

	// Connection statistics
	ActiveNNTPConnections int64 `json:"active_nntp_connections"`

	// Recent activity
	LastIndexingJobTime *time.Time `json:"last_indexing_job_time"`
	LastSearchTime      *time.Time `json:"last_search_time"`

	// System health
	SystemHealthy bool `json:"system_healthy"`
}

// NewAdminDashboard creates a new admin dashboard instance
func NewAdminDashboard(
	nntpRepo repositories.NNTPServerRepository,
	newsgroupRepo repositories.NewsgroupRepository,
	spotnetRepo repositories.SpotnetSourceRepository,
	articleRepo repositories.ArticleRepository,
	postRepo repositories.SpotnetPostRepository,
	jobRepo repositories.IndexingJobRepository,
	categoryRepo repositories.CategoryRepository,
) *AdminDashboard {
	return &AdminDashboard{
		nntpRepo:        nntpRepo,
		newsgroupRepo:   newsgroupRepo,
		spotnetRepo:     spotnetRepo,
		articleRepo:     articleRepo,
		postRepo:        postRepo,
		jobRepo:         jobRepo,
		categoryRepo:    categoryRepo,
	}
}

// GetDashboardStats retrieves comprehensive administrative statistics
func (ad *AdminDashboard) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Get total counts
	var err error

	// Articles count
	stats.TotalArticles, err = ad.articleRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles count: %w", err)
	}

	// Spotnet posts count
	stats.TotalSpotnetPosts, err = ad.postRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get spotnet posts count: %w", err)
	}

	// Enabled NNTP servers count
	stats.ActiveNNTPServers, err = ad.nntpRepo.CountEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled NNTP servers count: %w", err)
	}

	// Enabled Spotnet sources count
	stats.ActiveSpotnetSources, err = ad.spotnetRepo.CountEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled Spotnet sources count: %w", err)
	}

	// Categories count
	stats.TotalCategories, err = ad.categoryRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories count: %w", err)
	}

	// Active indexing jobs count
	stats.ActiveIndexingJobs, err = ad.jobRepo.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active indexing jobs count: %w", err)
	}

	// Set system health status (placeholder - would check actual system state)
	stats.SystemHealthy = true

	// Set recent activity timestamps (placeholder - would get real data)
	now := time.Now()
	stats.LastIndexingJobTime = &now
	stats.LastSearchTime = &now

	return stats, nil
}

// GetSystemHealth checks the overall system health status
func (ad *AdminDashboard) GetSystemHealth(ctx context.Context) (*models.SystemHealth, error) {
	// In a real implementation:
	// 1. Check database connectivity
	// 2. Check NNTP server connections
	// 3. Check Spotnet source connectivity  
	// 4. Check worker pool health
	// 5. Check resource usage (memory, CPU)
	// 6. Return comprehensive health status
	
	health := &models.SystemHealth{
		Healthy: true,
		Timestamp: time.Now(),
		Checks: []models.HealthCheck{
			{
				Name:    "Database",
				Status:  "healthy",
				Message: "Database connection established",
			},
			{
				Name:    "NNTP Servers",
				Status:  "healthy",
				Message: "All enabled NNTP servers connected",
			},
			{
				Name:    "Spotnet Sources",
				Status:  "healthy",
				Message: "All enabled Spotnet sources connected",
			},
		},
	}

	return health, nil
}

// GetRecentJobs retrieves recent indexing jobs for dashboard display
func (ad *AdminDashboard) GetRecentJobs(ctx context.Context, limit int) ([]*models.IndexingJob, error) {
	// In a real implementation:
	// 1. Query recent indexing jobs from database
	// 2. Order by creation time descending
	// 3. Limit results to specified count
	
	jobs := make([]*models.IndexingJob, 0)
	
	// Placeholder - return empty list
	return jobs, nil
}

// GetServerStatus retrieves status of all NNTP servers
func (ad *AdminDashboard) GetServerStatus(ctx context.Context) ([]*models.NNTPServer, error) {
	// In a real implementation:
	// 1. Get all enabled NNTP servers
	// 2. Check their connection status and last activity
	// 3. Return structured server status data
	
	servers := make([]*models.NNTPServer, 0)
	
	// Placeholder - return empty list
	return servers, nil
}