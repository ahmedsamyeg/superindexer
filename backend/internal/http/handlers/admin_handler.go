package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// AdminHandler handles admin endpoints.
type AdminHandler struct {
	DB   *bun.DB
	Log  interface{}
}

// GetDashboard handles getting dashboard statistics.
func (h *AdminHandler) GetDashboard(c echo.Context) error {
	ctx := c.Request().Context()
	var err error

	var totalArticles int64
	totalArticles, err = h.DB.NewSelect().Model((*Article)(nil)).Count(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get articles count: "+err.Error())
	}

	var totalSpotnet int64
	totalSpotnet, err = h.DB.NewSelect().Model((*SpotnetPost)(nil)).Count(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get spotnet count: "+err.Error())
	}

	var totalNewsgroups int64
	totalNewsgroups, err = h.DB.NewSelect().Model((*Newsgroup)(nil)).Count(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get newsgroups count: "+err.Error())
	}

	var enabledNewsgroups int64
	h.DB.NewSelect().Model((*Newsgroup)(nil)).Column("count(*)").Where("enabled = true").Scan(ctx, &enabledNewsgroups)

	var totalServers int64
	totalServers, err = h.DB.NewSelect().Model((*NNTPServer)(nil)).Count(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get servers count: "+err.Error())
	}

	var pendingJobs, runningJobs, failedJobs int64
	h.DB.NewSelect().Model((*IndexingJob)(nil)).Column("count(*)").Where("status = 'pending'").Scan(ctx, &pendingJobs)
	h.DB.NewSelect().Model((*IndexingJob)(nil)).Column("count(*)").Where("status = 'running'").Scan(ctx, &runningJobs)
	h.DB.NewSelect().Model((*IndexingJob)(nil)).Column("count(*)").Where("status = 'failed'").Scan(ctx, &failedJobs)

	var totalUsers int64
	totalUsers, err = h.DB.NewSelect().Model((*User)(nil)).Count(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get users count: "+err.Error())
	}

	var activeUsersCount int64
	h.DB.NewSelect().Model((*User)(nil)).Column("count(*)").Where("is_active = true").Scan(ctx, &activeUsersCount)

	var lastIndexedAt *time.Time
	err = h.DB.NewSelect().
		Model((*Article)(nil)).
		Column("max(posted_at)").
		Scan(ctx, &lastIndexedAt)
	if err != nil {
		lastIndexedAt = nil
	}

	dbSize, err := h.getDatabaseSize(ctx)
	if err != nil {
		dbSize = 0
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_articles":     totalArticles,
		"total_spotnet":      totalSpotnet,
		"total_newsgroups":   totalNewsgroups,
		"enabled_newsgroups": enabledNewsgroups,
		"total_servers":      totalServers,
		"pending_jobs":       pendingJobs,
		"running_jobs":       runningJobs,
		"failed_jobs":        failedJobs,
		"total_users":        totalUsers,
		"active_users":       activeUsersCount,
		"last_indexed":       lastIndexedAt,
		"db_size_bytes":      dbSize,
	})
}

// GetStats handles getting system statistics.
func (h *AdminHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()

	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)
	weekAgo := today.Add(-7 * 24 * time.Hour)
	monthAgo := today.Add(-30 * 24 * time.Hour)

	var articlesToday, articlesThisWeek, articlesThisMonth int64
	h.DB.NewSelect().Model((*Article)(nil)).Column("count(*)").Where("posted_at >= ?", today).Scan(ctx, &articlesToday)
	h.DB.NewSelect().Model((*Article)(nil)).Column("count(*)").Where("posted_at >= ?", weekAgo).Scan(ctx, &articlesThisWeek)
	h.DB.NewSelect().Model((*Article)(nil)).Column("count(*)").Where("posted_at >= ?", monthAgo).Scan(ctx, &articlesThisMonth)

	var spotnetToday, spotnetThisWeek, spotnetThisMonth int64
	h.DB.NewSelect().Model((*SpotnetPost)(nil)).Column("count(*)").Where("posted_at >= ?", today).Scan(ctx, &spotnetToday)
	h.DB.NewSelect().Model((*SpotnetPost)(nil)).Column("count(*)").Where("posted_at >= ?", weekAgo).Scan(ctx, &spotnetThisWeek)
	h.DB.NewSelect().Model((*SpotnetPost)(nil)).Column("count(*)").Where("posted_at >= ?", monthAgo).Scan(ctx, &spotnetThisMonth)

	var pendingJobs, runningJobs, completedJobs, failedJobs int64
	h.DB.NewSelect().Model((*IndexingJob)(nil)).Column("count(*)").Where("status = 'pending'").Scan(ctx, &pendingJobs)
	h.DB.NewSelect().Model((*IndexingJob)(nil)).Column("count(*)").Where("status = 'running'").Scan(ctx, &runningJobs)
	h.DB.NewSelect().Model((*IndexingJob)(nil)).Column("count(*)").Where("status = 'completed'").Scan(ctx, &completedJobs)
	h.DB.NewSelect().Model((*IndexingJob)(nil)).Column("count(*)").Where("status = 'failed'").Scan(ctx, &failedJobs)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"articles": map[string]interface{}{
			"total":      0,
			"today":      articlesToday,
			"this_week":  articlesThisWeek,
			"this_month": articlesThisMonth,
		},
		"spotnet": map[string]interface{}{
			"total":      0,
			"today":      spotnetToday,
			"this_week":  spotnetThisWeek,
			"this_month": spotnetThisMonth,
		},
		"jobs": map[string]interface{}{
			"pending":   pendingJobs,
			"running":   runningJobs,
			"completed": completedJobs,
			"failed":    failedJobs,
		},
	})
}

// GetSystemInfo handles getting system information.
func (h *AdminHandler) GetSystemInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"version":       "1.0.0",
		"go_version":    "go1.22",
		"os":            "linux",
		"arch":          "amd64",
		"uptime":        int64(0),
		"database":      "postgresql",
		"database_version": "16",
	})
}

// GetDBStats handles getting database statistics.
func (h *AdminHandler) GetDBStats(c echo.Context) error {
	ctx := c.Request().Context()

	var tables map[string]int64 = make(map[string]int64)

	tableNames := []string{"articles", "spotnet_posts", "newsgroups", "users", "nntp_servers", "spotnet_sources", "indexing_jobs"}
	for _, tableName := range tableNames {
		var count int64
		err := h.DB.NewRaw("SELECT count(*) FROM " + tableName).Scan(ctx, &count)
		if err == nil {
			tables[tableName] = count
		}
	}

	dbSize, err := h.getDatabaseSize(ctx)
	if err != nil {
		dbSize = 0
	}

	// Get connection pool stats via PostgreSQL system views
	var activeConns, idleConns, totalConns int64
	err = h.DB.NewRaw(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN state = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN state = 'idle' THEN 1 END) as idle
		FROM pg_stat_activity
		WHERE datname = current_database()
	`).Scan(ctx, &totalConns, &activeConns, &idleConns)
	if err != nil {
		activeConns = 0
		idleConns = 0
		totalConns = 0
	}

	// Get wait stats
	var waitCount int64
	err = h.DB.NewRaw(`
		SELECT count(*) FROM pg_stat_activity
		WHERE wait_event_type = 'Lock' AND wait_event IS NOT NULL
	`).Scan(ctx, &waitCount)
	if err != nil {
		waitCount = 0
	}

	// Get max connections setting
	var maxConns int64
	err = h.DB.NewRaw("SHOW max_connections").Scan(ctx, &maxConns)
	if err != nil {
		maxConns = 100
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tables": tables,
		"database_size_bytes": dbSize,
		"connections": map[string]interface{}{
			"max_open":        maxConns,
			"open":            totalConns,
			"in_use":          activeConns,
			"idle":            idleConns,
			"wait_count":      waitCount,
			"wait_duration":   "0s",
			"max_idle_closed": 0,
			"max_lifetime_closed": 0,
		},
	})
}

// getDatabaseSize returns the database size in bytes.
func (h *AdminHandler) getDatabaseSize(ctx context.Context) (int64, error) {
	var size int64
	err := h.DB.NewRaw("SELECT pg_database_size(current_database())").Scan(ctx, &size)
	return size, err
}


// Article is a local alias for models.Article
type Article = models.Article

// SpotnetPost is a local alias for models.SpotnetPost
type SpotnetPost = models.SpotnetPost

// Newsgroup is a local alias for models.Newsgroup
type Newsgroup = models.Newsgroup

// NNTPServer is a local alias for models.NNTPServer
type NNTPServer = models.NNTPServer

// IndexingJob is a local alias for models.IndexingJob
type IndexingJob = models.IndexingJob

// User is a local alias for models.User
type User = models.User
