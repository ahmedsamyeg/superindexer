package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/http/middleware"
	"github.com/superindexer/indexer/internal/http/service"
	"github.com/superindexer/indexer/internal/search"
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
	DB             *bun.DB
	Log            *slog.Logger
	AuthService    *service.AuthService
	UserService    *service.UserService
	SearchEngine   *search.SearchEngine
	JWTSecret      string
	RefreshTokens  *service.RefreshTokenService
	TokenBlacklist *service.TokenBlacklistService
}

// AuthMiddleware returns the auth middleware function.
func (h *Handlers) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return middleware.AuthMiddleware(h.DB, h.JWTSecret)(next)
}

// AdminMiddleware returns the admin middleware function.
func (h *Handlers) AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return middleware.AdminMiddleware()(next)
}

// NewHandlers creates all handlers with shared dependencies.
func NewHandlers(db *bun.DB, log *slog.Logger, jwtSecret string) *Handlers {
	// Create repositories
	userRepo := repositories.NewUserRepository(db)
	_ = repositories.NewArticleRepository(db)
	_ = repositories.NewSearchRepository(db)
	nntpRepo := repositories.NewNNTPServerRepository(db)
	newsgroupRepo := repositories.NewNewsgroupRepository(db)
	spotnetSourceRepo := repositories.NewSpotnetSourceRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	indexingRepo := repositories.NewIndexingJobRepository(db)

	// Create auth services
	refreshTokens := service.NewRefreshTokenService(db, userRepo, jwtSecret, 24*60*60)
	tokenBlacklist := service.NewTokenBlacklistService(db)

	// Create handlers
	return &Handlers{
		AuthHandler:          &AuthHandler{AuthService: nil, DB: db, Log: log, JWTSecret: jwtSecret, RefreshTokens: refreshTokens, TokenBlacklist: tokenBlacklist},
		UserHandler:          &UserHandler{Log: log, UserService: service.NewUserService(userRepo, service.DefaultAuthConfig())},
		SearchHandler:        &SearchHandler{Log: log, SearchEngine: &search.SearchEngine{DB: db}},
		SearchHistoryHandler: &SearchHistoryHandler{Log: log, DB: db},
		BookmarkHandler:      &BookmarkHandler{Log: log, DB: db},
		NNTPServerHandler:    &NNTPServerHandler{Log: log, DB: db, Repo: nntpRepo},
		NewsgroupHandler:     &NewsgroupHandler{Log: log, DB: db, Repo: newsgroupRepo},
		CategoryHandler:      &CategoryHandler{Log: log, DB: db, Repo: categoryRepo},
		SpotnetSourceHandler: &SpotnetSourceHandler{Log: log, DB: db, Repo: spotnetSourceRepo},
		SpotnetPostHandler:   &SpotnetPostHandler{Log: log, DB: db, Repo: spotnetSourceRepo},
		IndexingJobHandler:   &IndexingJobHandler{Log: log, DB: db, Repo: indexingRepo},
		AdminHandler:         &AdminHandler{Log: log, DB: db},
		SettingsHandler:      &SettingsHandler{Log: log, DB: db},
		DB:                   db,
		Log:                  log,
		JWTSecret:            jwtSecret,
		RefreshTokens:        refreshTokens,
		TokenBlacklist:       tokenBlacklist,
	}
}
