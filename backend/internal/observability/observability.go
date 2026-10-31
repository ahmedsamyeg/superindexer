package observability

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Observability holds observability-related components.
type Observability struct {
	metrics   *Metrics
	tracer    interface{}
	mux       *http.ServeMux
	promHandler http.Handler
	log       *slog.Logger
	once      sync.Once
}

// Metrics holds Prometheus metrics.
type Metrics struct {
	ArticlesProcessedTotal     *prometheus.CounterVec
	SpotnetPostsProcessedTotal *prometheus.CounterVec
	ProcessingErrorsTotal      *prometheus.CounterVec
	JobsActive                 *prometheus.GaugeVec
	JobsFailedTotal            *prometheus.CounterVec
	WorkerUtilization          *prometheus.GaugeVec
	NNTPConnectionsActive      prometheus.Gauge
	IngestionDuration          *prometheus.HistogramVec
	SearchDuration             *prometheus.HistogramVec
	DatabaseConnectionsActive  prometheus.Gauge
}

// New creates a new Observability instance.
func New(log *slog.Logger) *Observability {
	m := &Observability{
		log: log,
		mux: http.NewServeMux(),
	}

	m.metrics = &Metrics{
		ArticlesProcessedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "indexer_articles_processed_total",
				Help: "Total processed articles",
			},
			[]string{"source", "status"},
		),
		SpotnetPostsProcessedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "indexer_spotnet_posts_processed_total",
				Help: "Total processed Spotnet posts",
			},
			[]string{"source", "status"},
		),
		ProcessingErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "indexer_processing_errors_total",
				Help: "Processing failures",
			},
			[]string{"component", "error_type"},
		),
		JobsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "indexer_jobs_active",
				Help: "Active indexing jobs",
			},
			[]string{"job_type"},
		),
		JobsFailedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "indexer_jobs_failed_total",
				Help: "Failed jobs",
			},
			nil,
		),
		WorkerUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "indexer_worker_utilization",
				Help: "Worker utilization percentage",
			},
			[]string{"worker_id"},
		),
		NNTPConnectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "indexer_nntp_connections_active",
			Help: "Active NNTP connections",
		}),
		IngestionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "indexer_ingestion_duration_seconds",
				Help:    "Processing duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"source_type", "status"},
		),
		SearchDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "indexer_search_duration_seconds",
				Help:    "Search execution time",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query_type"},
		),
		DatabaseConnectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "indexer_database_connections_active",
			Help: "Active database connections",
		}),
	}

	return m
}

// Start initializes and registers all metrics.
func (o *Observability) Start(ctx context.Context) error {
	o.once.Do(func() {
		// Register metrics with Prometheus
		prometheus.MustRegister(o.metrics.ArticlesProcessedTotal)
		prometheus.MustRegister(o.metrics.SpotnetPostsProcessedTotal)
		prometheus.MustRegister(o.metrics.ProcessingErrorsTotal)
		prometheus.MustRegister(o.metrics.JobsActive)
		prometheus.MustRegister(o.metrics.JobsFailedTotal)
		prometheus.MustRegister(o.metrics.WorkerUtilization)
		prometheus.MustRegister(o.metrics.NNTPConnectionsActive)
		prometheus.MustRegister(o.metrics.IngestionDuration)
		prometheus.MustRegister(o.metrics.SearchDuration)
		prometheus.MustRegister(o.metrics.DatabaseConnectionsActive)

		// Register Prometheus HTTP handler
		o.promHandler = promhttp.Handler()
		o.mux.Handle("/metrics", o.promHandler)

		// Initialize OpenTelemetry
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

		o.log.Info("observability initialized")
	})

	return nil
}

// Stop cleans up observability resources.
func (o *Observability) Stop(ctx context.Context) error {
	// TODO: Implement proper tracer shutdown
	return nil
}

// RegisterEchoMiddleware registers the Prometheus middleware with Echo.
func (o *Observability) RegisterEchoMiddleware(e *echo.Echo) {
	// Prometheus metrics middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start).Seconds()

			// Record metrics
			o.metrics.SearchDuration.WithLabelValues("http").Observe(duration)

			return err
		}
	})
}

// GetMetrics returns the Prometheus metrics.
func (o *Observability) GetMetrics() *Metrics {
	return o.metrics
}

// GetPrometheusHandler returns the Prometheus HTTP handler.
func (o *Observability) GetPrometheusHandler() http.Handler {
	return o.promHandler
}

// NewLogger creates a structured logger.
func NewLogger(level string, format string) *slog.Logger {
	var opts slog.HandlerOptions
	switch level {
	case "debug":
		opts = slog.HandlerOptions{Level: slog.LevelDebug}
	case "warn":
		opts = slog.HandlerOptions{Level: slog.LevelWarn}
	case "error":
		opts = slog.HandlerOptions{Level: slog.LevelError}
	default:
		opts = slog.HandlerOptions{Level: slog.LevelInfo}
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &opts)
	}

	return slog.New(handler)
}
