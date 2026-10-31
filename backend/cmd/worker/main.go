package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/superindexer/indexer/internal/config"
	"github.com/superindexer/indexer/internal/database"
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

	log.Info("starting indexer worker service")

	// Initialize database
	db, err := database.New(context.Background(), database.Config{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	}, log)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("database initialized successfully")

	// Initialize observability
	observability := observability.New(log)
	if err := observability.Start(context.Background()); err != nil {
		log.Error("failed to start observability", "error", err)
	}
	defer observability.Stop(context.Background())

	// TODO: Initialize indexing workers
	// TODO: Start NNTP ingestion
	// TODO: Start Spotnet ingestion
	// TODO: Start job scheduler

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Info("received shutdown signal")
		cancel()
	}()

	log.Info("worker service started, waiting for jobs...")

	// Block until context is cancelled
	<-ctx.Done()

	log.Info("worker service stopped")
}
