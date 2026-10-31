package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/superindexer/indexer/internal/config"
	"github.com/superindexer/indexer/internal/http/handlers"
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
) *Server {
	e := echo.New()

	s := &Server{
		echo:     e,
		cfg:      cfg,
		log:      log,
		handlers: handlers.NewHandlers(db, log),
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return s
}

// setupMiddleware configures all middleware.
func (s *Server) setupMiddleware() {
	// Request ID middleware
	s.echo.Use(middleware.RequestID())

	// Recovery middleware
	s.echo.Use(middleware.Recover())

	// Logger middleware
	s.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			if path == "/health" || path == "/ready" {
				return true
			}
			return false
		},
	}))

	// CORS middleware
	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders: []string{"X-Request-ID", "X-Total-Count"},
		MaxAge:       86400,
	}))

	// Secure headers middleware
	s.echo.Use(middleware.Secure())

	// Request body size limit
	s.echo.Use(middleware.BodyLimit("10M"))
}

// setupRoutes configures all API routes.
func (s *Server) setupRoutes() {
	api := s.echo.Group("/api")
	v1 := api.Group("/v1")

	// Health endpoints (no auth required)
	s.echo.GET("/health", s.handlers.HealthHandler)
	s.echo.GET("/ready", s.handlers.ReadyHandler)

	// Auth endpoints
	auth := v1.Group("/auth")
	auth.POST("/login", s.handlers.AuthHandler.Login)
	auth.POST("/register", s.handlers.AuthHandler.Register)
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
	users.PATCH("/:id/toggle", s.handlers.UserHandler.ToggleActive)

	// Search endpoints
	search := v1.Group("/search")
	search.GET("", s.handlers.SearchHandler.Search)
	search.GET("/articles", s.handlers.SearchHandler.SearchArticles)
	search.GET("/spotnet", s.handlers.SearchHandler.SearchSpotnet)
	search.GET("/facets", s.handlers.SearchHandler.GetFacets)

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

	// NNTP Server management
	servers := v1.Group("/servers")
	servers.Use(s.handlers.AuthMiddleware)
	servers.Use(s.handlers.AdminMiddleware)
	servers.GET("", s.handlers.NNTPServerHandler.List)
	servers.GET("/:id", s.handlers.NNTPServerHandler.GetByID)
	servers.POST("", s.handlers.NNTPServerHandler.Create)
	servers.PUT("/:id", s.handlers.NNTPServerHandler.Update)
	servers.DELETE("/:id", s.handlers.NNTPServerHandler.Delete)
	servers.POST("/:id/test", s.handlers.NNTPServerHandler.TestConnection)

	// Newsgroup management
	groups := v1.Group("/groups")
	groups.Use(s.handlers.AuthMiddleware)
	groups.Use(s.handlers.AdminMiddleware)
	groups.GET("", s.handlers.NewsgroupHandler.List)
	groups.GET("/:id", s.handlers.NewsgroupHandler.GetByID)
	groups.POST("", s.handlers.NewsgroupHandler.Create)
	groups.PUT("/:id", s.handlers.NewsgroupHandler.Update)
	groups.DELETE("/:id", s.handlers.NewsgroupHandler.Delete)
	groups.GET("/:id/discover", s.handlers.NewsgroupHandler.Discover)
	groups.POST("/:id/index", s.handlers.NewsgroupHandler.ManualIndex)

	// Categories
	categories := v1.Group("/categories")
	categories.Use(s.handlers.AuthMiddleware)
	categories.GET("", s.handlers.CategoryHandler.List)
	categories.GET("/:id", s.handlers.CategoryHandler.GetByID)
	categories.POST("", s.handlers.CategoryHandler.Create)
	categories.PUT("/:id", s.handlers.CategoryHandler.Update)
	categories.DELETE("/:id", s.handlers.CategoryHandler.Delete)

	// Spotnet sources
	sources := v1.Group("/spotnet/sources")
	sources.Use(s.handlers.AuthMiddleware)
	sources.Use(s.handlers.AdminMiddleware)
	sources.GET("", s.handlers.SpotnetSourceHandler.List)
	sources.GET("/:id", s.handlers.SpotnetSourceHandler.GetByID)
	sources.POST("", s.handlers.SpotnetSourceHandler.Create)
	sources.PUT("/:id", s.handlers.SpotnetSourceHandler.Update)
	sources.DELETE("/:id", s.handlers.SpotnetSourceHandler.Delete)
	sources.POST("/:id/test", s.handlers.SpotnetSourceHandler.TestConnection)
	sources.POST("/:id/sync", s.handlers.SpotnetSourceHandler.Sync)

	// Spotnet posts
	posts := v1.Group("/spotnet/posts")
	posts.Use(s.handlers.AuthMiddleware)
	posts.GET("", s.handlers.SpotnetPostHandler.List)
	posts.GET("/:id", s.handlers.SpotnetPostHandler.GetByID)

	// Indexing jobs
	jobs := v1.Group("/indexing/jobs")
	jobs.Use(s.handlers.AuthMiddleware)
	jobs.GET("", s.handlers.IndexingJobHandler.List)
	jobs.GET("/:id", s.handlers.IndexingJobHandler.GetByID)
	jobs.POST("", s.handlers.IndexingJobHandler.Create)
	jobs.POST("/:id/pause", s.handlers.IndexingJobHandler.Pause)
	jobs.POST("/:id/resume", s.handlers.IndexingJobHandler.Resume)
	jobs.POST("/:id/retry", s.handlers.IndexingJobHandler.Retry)
	jobs.DELETE("/:id", s.handlers.IndexingJobHandler.Cancel)

	// Admin endpoints
	admin := v1.Group("/admin")
	admin.Use(s.handlers.AuthMiddleware)
	admin.Use(s.handlers.AdminMiddleware)
	admin.GET("/dashboard", s.handlers.AdminHandler.GetDashboard)
	admin.GET("/stats", s.handlers.AdminHandler.GetStats)
	admin.GET("/system", s.handlers.AdminHandler.GetSystemInfo)
	admin.GET("/logs", s.handlers.AdminHandler.GetLogs)
	admin.GET("/db/stats", s.handlers.AdminHandler.GetDBStats)

	// Settings
	settings := v1.Group("/settings")
	settings.Use(s.handlers.AuthMiddleware)
	settings.Use(s.handlers.AdminMiddleware)
	settings.GET("", s.handlers.SettingsHandler.List)
	settings.GET("/:key", s.handlers.SettingsHandler.Get)
	settings.PUT("/:key", s.handlers.SettingsHandler.Update)

	// Static frontend files (SPA fallback)
	s.echo.Static("/static", "frontend/dist/static")
	s.echo.GET("/*", func(c echo.Context) error {
		return c.File("frontend/dist/index.html")
	})
}

// Start begins serving HTTP requests.
func (s *Server) Start(ctx context.Context) error {
	s.log.Info("starting HTTP server", "addr", s.server.Addr)

	go func() {
		<-ctx.Done()
		s.log.Info("shutting down HTTP server")
	}()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("initiating HTTP server shutdown")
	return s.server.Shutdown(ctx)
}
