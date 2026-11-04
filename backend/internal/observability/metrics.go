package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector collects and exposes application metrics
type MetricsCollector struct {
	// Application metrics
	apiRequestsTotal *prometheus.CounterVec
	searchDuration   prometheus.Histogram
	indexingDuration prometheus.Histogram
	dbConnections    prometheus.Gauge

	// NNTP server metrics
	nntpConnectionsActive *prometheus.GaugeVec
	nntpServersEnabled    prometheus.Gauge

	// Spotnet metrics
	spotnetSourcesEnabled prometheus.Gauge
	spotnetPostsProcessed *prometheus.CounterVec

	// Worker metrics
	workerPoolSize   prometheus.Gauge
	workerJobsActive prometheus.Gauge

	// Application health metrics
	appHealthy prometheus.Gauge
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		apiRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "indexer_api_requests_total",
				Help: "Total number of API requests by endpoint and status",
			},
			[]string{"endpoint", "status"},
		),
		searchDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name: "indexer_search_duration_seconds",
				Help: "Duration of search operations in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
			},
		),
		indexingDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name: "indexer_indexing_duration_seconds",
				Help: "Duration of indexing operations in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60},
			},
		),
		dbConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "indexer_database_connections_active",
				Help: "Active database connections",
			},
		),
		nntpConnectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "indexer_nntp_connections_active",
				Help: "Active NNTP connections per server",
			},
			[]string{"server_id"},
		),
		nntpServersEnabled: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "indexer_nntp_servers_enabled",
				Help: "Number of enabled NNTP servers",
			},
		),
		spotnetSourcesEnabled: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "indexer_spotnet_sources_enabled",
				Help: "Number of enabled Spotnet sources",
			},
		),
		spotnetPostsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "indexer_spotnet_posts_processed_total",
				Help: "Total number of Spotnet posts processed",
			},
			[]string{"source_id"},
		),
		workerPoolSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "indexer_worker_pool_size",
				Help: "Number of workers in the pool",
			},
		),
		workerJobsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "indexer_worker_jobs_active",
				Help: "Number of active indexing jobs being processed",
			},
		),
		appHealthy: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "indexer_app_healthy",
				Help: "Application health status (1 = healthy, 0 = unhealthy)",
			},
		),
	}
}

// CollectMetrics collects and updates all metrics
func (mc *MetricsCollector) CollectMetrics(ctx context.Context, appState map[string]interface{}) {
	// Update application health status
	if appState["healthy"] == true {
		mc.appHealthy.Set(1)
	} else {
		mc.appHealthy.Set(0)
	}

	// Update database connection count (placeholder)
	mc.dbConnections.Set(float64(0)) // This would come from actual connection pool metrics

	// Update worker pool size
	if workers, ok := appState["workers"]; ok {
		if workerCount, ok := workers.(int); ok {
			mc.workerPoolSize.Set(float64(workerCount))
		}
	}

	// Update job count
	if jobs, ok := appState["active_jobs"]; ok {
		if jobCount, ok := jobs.(int); ok {
			mc.workerJobsActive.Set(float64(jobCount))
		}
	}

	// Update server counts
	if servers, ok := appState["nntp_servers"]; ok {
		if serverCount, ok := servers.(int); ok {
			mc.nntpServersEnabled.Set(float64(serverCount))
		}
	}

	if sources, ok := appState["spotnet_sources"]; ok {
		if sourceCount, ok := sources.(int); ok {
			mc.spotnetSourcesEnabled.Set(float64(sourceCount))
		}
	}
}

// RecordAPICall records an API call
func (mc *MetricsCollector) RecordAPICall(endpoint, status string) {
	mc.apiRequestsTotal.WithLabelValues(endpoint, status).Inc()
}

// RecordSearchDuration records the duration of a search operation
func (mc *MetricsCollector) RecordSearchDuration(duration time.Duration) {
	mc.searchDuration.Observe(duration.Seconds())
}

// RecordIndexingDuration records the duration of an indexing operation
func (mc *MetricsCollector) RecordIndexingDuration(duration time.Duration) {
	mc.indexingDuration.Observe(duration.Seconds())
}

// RecordSpotnetPostProcessed records a processed Spotnet post
func (mc *MetricsCollector) RecordSpotnetPostProcessed(sourceID int64) {
	mc.spotnetPostsProcessed.WithLabelValues(fmt.Sprintf("%d", sourceID)).Inc()
}

// RecordActiveNNTPConnection records an active NNTP connection
func (mc *MetricsCollector) RecordActiveNNTPConnection(serverID int64) {
	mc.nntpConnectionsActive.WithLabelValues(fmt.Sprintf("%d", serverID)).Add(1)
}

// RecordInactiveNNTPConnection decrements an active NNTP connection count
func (mc *MetricsCollector) RecordInactiveNNTPConnection(serverID int64) {
	mc.nntpConnectionsActive.WithLabelValues(fmt.Sprintf("%d", serverID)).Sub(1)
}

// GetMetricsHandler returns an HTTP handler for Prometheus metrics
func (mc *MetricsCollector) GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// HealthCheck checks if the application is healthy
func (mc *MetricsCollector) HealthCheck(ctx context.Context) bool {
	// In a real implementation, this would check database connectivity,
	// worker pool health, and other critical system components
	return true
}