package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/http/middleware"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/uptrace/bun"
)

// Handlers holds all HTTP handlers.
type Handlers struct {
	AuthHandler          *AuthHandler
	UserHandler          *UserHandler
	SearchHandler        *SearchHandler
	SearchHistoryHandler *SearchHistoryHandler
	BookmarkHandler      *BookmarkHandler
	NNTPServerHandler    *NNTPServerHandler
	NewsgroupHandler     *NewsgroupHandler
	CategoryHandler      *CategoryHandler
	SpotnetSourceHandler *SpotnetSourceHandler
	SpotnetPostHandler   *SpotnetPostHandler
	IndexingJobHandler   *IndexingJobHandler
	AdminHandler         *AdminHandler
	SettingsHandler      *SettingsHandler

	// Shared dependencies
	DB    *bun.DB
	Log   *slog.Logger
	AuthService *service.AuthService
	UserService *service.UserService
}

// AuthMiddleware returns the auth middleware function.
func (h *Handlers) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return middleware.AuthMiddleware()(next)
}

// AdminMiddleware returns the admin middleware function.
func (h *Handlers) AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return middleware.AdminMiddleware()(next)
}

// NewHandlers creates all handlers with shared dependencies.
func NewHandlers(db *bun.DB, log *slog.Logger, authService *service.AuthService, userService *service.UserService) *Handlers {
	return &Handlers{
		AuthHandler:          &AuthHandler{AuthService: authService, Log: log},
		UserHandler:          &UserHandler{Log: log, UserService: userService},
		SearchHandler:        &SearchHandler{Log: log},
		SearchHistoryHandler: &SearchHistoryHandler{Log: log},
		BookmarkHandler:      &SearchHistoryHandler{Log: log}, // Placeholder - should be updated
		NNTPServerHandler:    &NNTPServerHandler{Log: log},
		NewsgroupHandler:     &NewsgroupHandler{Log: log},
		CategoryHandler:      &CategoryHandler{Log: log},
		SpotnetSourceHandler: &SpotnetSourceHandler{Log: log},
		SpotnetPostHandler:   &SpotnetPostHandler{Log: log},
		IndexingJobHandler:   &IndexingJobHandler{Log: log},
		AdminHandler:         &AdminHandler{Log: log},
		SettingsHandler:      &SettingsHandler{Log: log},
		DB:    db,
		Log:   log,
		AuthService: authService,
		UserService: userService,
	}
}