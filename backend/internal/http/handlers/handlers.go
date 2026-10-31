package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/http/middleware"
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
func NewHandlers(db *bun.DB, log *slog.Logger) *Handlers {
	return &Handlers{
		AuthHandler:          &AuthHandler{DB: db, Log: log},
		UserHandler:          &UserHandler{DB: db, Log: log},
		SearchHandler:        &SearchHandler{DB: db, Log: log},
		SearchHistoryHandler: &SearchHistoryHandler{DB: db, Log: log},
		BookmarkHandler:      &BookmarkHandler{DB: db, Log: log},
		NNTPServerHandler:    &NNTPServerHandler{DB: db, Log: log},
		NewsgroupHandler:     &NewsgroupHandler{DB: db, Log: log},
		CategoryHandler:      &CategoryHandler{DB: db, Log: log},
		SpotnetSourceHandler: &SpotnetSourceHandler{DB: db, Log: log},
		SpotnetPostHandler:   &SpotnetPostHandler{DB: db, Log: log},
		IndexingJobHandler:   &IndexingJobHandler{DB: db, Log: log},
		AdminHandler:         &AdminHandler{DB: db, Log: log},
		SettingsHandler:      &SettingsHandler{DB: db, Log: log},
		DB:    db,
		Log:   log,
	}
}
