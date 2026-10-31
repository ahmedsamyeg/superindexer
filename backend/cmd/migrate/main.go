package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/superindexer/indexer/internal/config"
	"github.com/superindexer/indexer/internal/observability"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create logger
	log := observability.NewLogger(cfg.Observability.LogLevel, cfg.Observability.LogFormat)

	dsn := cfg.Database.DSN
	if dsn == "" {
		dsn = cfg.Database.GetDSN()
	}
	logDir := "internal/database/migrations"

	absDir, err := filepath.Abs(logDir)
	if err != nil {
		log.Error("failed to get absolute path for migrations", "error", err)
		os.Exit(1)
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", absDir), dsn)
	if err != nil {
		log.Error("failed to create migrate instance", "error", err)
		os.Exit(1)
	}
	defer m.Close()

	// Get subcommand
	subcmd := "up"
	if len(os.Args) > 1 {
		subcmd = os.Args[1]
	}

	switch subcmd {
	case "up":
		log.Info("running migrations up")
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Info("no changes to migrate")
			} else {
				log.Error("failed to run migrations up", "error", err)
				os.Exit(1)
			}
		}
		log.Info("migrations applied successfully")

	case "down":
		log.Info("running migrations down")
		if err := m.Steps(-1); err != nil {
			if err == migrate.ErrNoChange {
				log.Info("no changes to migrate")
			} else {
				log.Error("failed to run migrations down", "error", err)
				os.Exit(1)
			}
		}
		log.Info("migration rolled back successfully")

	case "force":
		if len(os.Args) < 3 {
			log.Error("force requires a version number")
			os.Exit(1)
		}
		version := os.Args[2]
		log.Info("forcing migration to version", "version", version)
		if err := m.Force(1); err != nil {
			log.Error("failed to force migration", "error", err)
			os.Exit(1)
		}
		log.Info("migration version forced successfully")

	case "status":
		version, dirty, err := m.Version()
		if err != nil {
			log.Error("failed to get migration status", "error", err)
			os.Exit(1)
		}
		log.Info("current migration status", "version", version, "dirty", dirty)

	default:
		log.Error("unknown command", "command", subcmd)
		fmt.Println("Usage: migrate [up|down|force <version>|status]")
		os.Exit(1)
	}
}
