package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/superindexer/indexer/internal/config"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/handlers"
	appmiddleware "github.com/superindexer/indexer/internal/http/middleware"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/uptrace/bun"
)

// Server wraps the Echo HTTP server with application services.
type Server struct {
	echo     *echo.Echo
	server   *http.Server
	cfg      *config.Config
	log      *slog.Logger
	handlers *handlers.Handlers
}

// New creates a new HTTP server.
func New(
	cfg *config.Config,
	log *slog.Logger,
	db *bun.DB,
	jwtSecret string,
) (*Server, error) {
	e := echo.New()

	s := &Server{
		echo:     e,
		cfg:      cfg,
		log:      log,
		handlers: handlers.NewHandlers(db, log, jwtSecret),
	}

	s.setupServices()
	s.setupMiddleware()
	s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return s, nil
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	s.log.Info("starting HTTP server", 
		slog.String("address", s.server.Addr),
		slog.Int("port", s.cfg.Server.Port))
	
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// setupServices initializes shared services.
func (s *Server) setupServices() {
	// Create email service
	emailService := service.NewEmailService(service.EmailConfig{
		Host:    s.cfg.Email.Host,
		Port:    s.cfg.Email.Port,
		Username: s.cfg.Email.Username,
		Password: s.cfg.Email.Password,
		From:     s.cfg.Email.From,
		Enabled:  s.cfg.Email.Enabled,
	})

	// Initialize AuthService with all dependencies
	s.handlers.AuthService = service.NewAuthService(
		repositories.NewUserRepository(s.handlers.DB),
		s.cfg.Auth,
		emailService,
		s.handlers.RefreshTokens,
		s.handlers.TokenBlacklist,
	)
}

// setupMiddleware configures all appmiddleware.
func (s *Server) setupMiddleware() {
	// Request ID middleware
	s.echo.Use(appmiddleware.RequestID())

	// Recovery middleware
	s.echo.Use(appmiddleware.Recover())

	// Logger middleware
	s.echo.Use(appmiddleware.LoggerWithConfig(appmiddleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			if path == "/health" || path == "/ready" {
				return true
			}
			return false
		},
	}))

	// CORS middleware
	s.echo.Use(appmiddleware.CORSWithConfig(appmiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	// Timeout middleware
	s.echo.Use(appmiddleware.TimeoutWithConfig(appmiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
}

// setupRoutes configures all API routes.
func (s *Server) setupRoutes() {
	// Health check endpoints
	s.echo.GET("/health", s.handlers.HealthCheck)
	s.echo.GET("/ready", s.handlers.ReadyCheck)

	// Prometheus metrics (no auth required)
	s.echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Version endpoint
	s.echo.GET("/version", s.handlers.Version)

	// API v1 routes
	v1 := s.echo.Group("/api/v1")
	
	// Authentication endpoints with rate limiting
	auth := v1.Group("/auth")
	auth.Use(appmiddleware.RateLimitMiddleware(5, 1*time.Minute)) // 5 requests per minute
	auth.POST("/login", s.handlers.AuthHandler.Login)
	auth.POST("/register", s.handlers.AuthHandler.Register)
	auth.POST("/refresh", s.handlers.AuthHandler.Refresh)
	auth.POST("/logout", s.handlers.AuthMiddleware(s.handlers.AuthHandler.Logout))
	auth.GET("/me", s.handlers.AuthMiddleware(s.handlers.AuthHandler.GetMe))
	auth.POST("/password/reset", s.handlers.AuthHandler.RequestPasswordReset)
	auth.POST("/password/reset/confirm", s.handlers.AuthHandler.ConfirmPasswordReset)

	// User management
	users := v1.Group("/users")
	users.Use(s.handlers.AuthMiddleware)
	users.Use(s.handlers.AdminMiddleware)
	users.GET("", s.handlers.UserHandler.List)
	users.GET("/:id", s.handlers.UserHandler.GetByID)
	users.POST("", s.handlers.UserHandler.Create)
	users.PUT("/:id", s.handlers.UserHandler.Update)
	users.DELETE("/:id", s.handlers.UserHandler.Delete)
	users.POST("/:id/toggle-active", s.handlers.UserHandler.ToggleActive)

	// Search endpoints
	search := v1.Group("/search")
	search.Use(s.handlers.AuthMiddleware)
	search.GET("", s.handlers.SearchHandler.Search)
	search.GET("/articles", s.handlers.SearchHandler.SearchArticles)
	search.GET("/spotnet", s.handlers.SearchHandler.SearchSpotnet)
	search.GET("/categories", s.handlers.SearchHandler.GetCategories)
	search.GET("/facets", s.handlers.SearchHandler.GetFacets)
	search.GET("/suggestions", s.handlers.SearchHandler.GetSuggestions)

	// NNTP Server management
	servers := v1.Group("/servers")
	servers.Use(s.handlers.AuthMiddleware)
	servers.Use(s.handlers.AdminMiddleware)
	servers.GET("", s.handlers.NNTPServerHandler.List)
	servers.GET("/:id", s.handlers.NNTPServerHandler.GetByID)
	servers.POST("", s.handlers.NNTPServerHandler.Create)
	servers.PUT("/:id", s.handlers.NNTPServerHandler.Update)
	servers.DELETE("/:id", s.handlers.NNTPServerHandler.Delete)

	// Newsgroup management
	groups := v1.Group("/groups")
	groups.Use(s.handlers.AuthMiddleware)
	groups.Use(s.handlers.AdminMiddleware)
	groups.GET("", s.handlers.NewsgroupHandler.List)
	groups.GET("/:id", s.handlers.NewsgroupHandler.GetByID)
	groups.POST("", s.handlers.NewsgroupHandler.Create)
	groups.PUT("/:id", s.handlers.NewsgroupHandler.Update)
	groups.DELETE("/:id", s.handlers.NewsgroupHandler.Delete)

	// Spotnet sources
	sources := v1.Group("/spotnet/sources")
	sources.Use(s.handlers.AuthMiddleware)
	sources.Use(s.handlers.AdminMiddleware)
	sources.GET("", s.handlers.SpotnetSourceHandler.List)
	sources.GET("/:id", s.handlers.SpotnetSourceHandler.GetByID)
	sources.POST("", s.handlers.SpotnetSourceHandler.Create)
	sources.PUT("/:id", s.handlers.SpotnetSourceHandler.Update)
	sources.DELETE("/:id", s.handlers.SpotnetSourceHandler.Delete)

	// Indexing jobs
	jobs := v1.Group("/indexing/jobs")
	jobs.Use(s.handlers.AuthMiddleware)
	jobs.Use(s.handlers.AdminMiddleware)
	jobs.GET("", s.handlers.IndexingJobHandler.List)
	jobs.GET("/:id", s.handlers.IndexingJobHandler.GetByID)
	jobs.POST("", s.handlers.IndexingJobHandler.Create)
	jobs.PUT("/:id/retry", s.handlers.IndexingJobHandler.Retry)
	jobs.PUT("/:id/cancel", s.handlers.IndexingJobHandler.Cancel)

	// Admin endpoints
	admin := v1.Group("/admin")
	admin.Use(s.handlers.AuthMiddleware)
	admin.Use(s.handlers.AdminMiddleware)
	admin.GET("/dashboard", s.handlers.AdminHandler.GetDashboard)
	admin.GET("/stats", s.handlers.AdminHandler.GetStats)
	admin.GET("/system-info", s.handlers.AdminHandler.GetSystemInfo)
	admin.GET("/db-stats", s.handlers.AdminHandler.GetDBStats)

	// Settings
	settings := v1.Group("/settings")
	settings.Use(s.handlers.AuthMiddleware)
	settings.Use(s.handlers.AdminMiddleware)
	settings.GET("", s.handlers.SettingsHandler.List)
	settings.GET("/:key", s.handlers.SettingsHandler.Get)
	settings.PUT("/:key", s.handlers.SettingsHandler.Update)

	// Bookmarks
	bookmarks := v1.Group("/bookmarks")
	bookmarks.Use(s.handlers.AuthMiddleware)
	bookmarks.GET("", s.handlers.BookmarkHandler.List)
	bookmarks.POST("", s.handlers.BookmarkHandler.Create)
	bookmarks.DELETE("/:id", s.handlers.BookmarkHandler.Delete)

	// Search history
	history := v1.Group("/history")
	history.Use(s.handlers.AuthMiddleware)
	history.GET("", s.handlers.SearchHistoryHandler.List)
	history.POST("", s.handlers.SearchHistoryHandler.Create)
	history.DELETE("", s.handlers.SearchHistoryHandler.Clear)

	// Categories
	categories := v1.Group("/categories")
	categories.Use(s.handlers.AuthMiddleware)
	categories.Use(s.handlers.AdminMiddleware)
	categories.GET("", s.handlers.CategoryHandler.List)
	categories.GET("/:id", s.handlers.CategoryHandler.GetByID)
	categories.POST("", s.handlers.CategoryHandler.Create)
	categories.PUT("/:id", s.handlers.CategoryHandler.Update)
	categories.DELETE("/:id", s.handlers.CategoryHandler.Delete)
}