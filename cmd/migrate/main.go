package main

import (
    "context"
    "database/sql"
    "fmt"
    "log/slog"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/stdlib"
    "github.com/uptrace/bun/dialect/pgdialect"
    "github.com/uptrace/bun"

    "github.com/superindexer/indexer/internal/database/migrations"
)

func main() {
    ctx := context.Background()
    dsn := os.Getenv("DATABASE_DSN")
    if dsn == "" {
        slog.Error("DATABASE_DSN not set")
        os.Exit(1)
    }

    // Init pool similar to backend/database
    pool, sqldb, err := initPool(ctx, dsn)
    if err != nil {
        slog.Error("init pool", "err", err)
        os.Exit(1)
    }
    defer pool.Close()

    // Run migrations
    r := migrations.NewRunner("internal/database/migrations")
    if err := r.Run(ctx, sqldb); err != nil {
        slog.Error("migration failed", "err", err)
        os.Exit(1)
    }
    slog.Info("migrations applied successfully")
}

func initPool(ctx context.Context, dsn string) (*pgxpool.Pool, *sql.DB, error) {
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, nil, fmt.Errorf("parse DSN: %w", err)
    }
    pool, err := pgxpool.NewWithConfig(ctx, cfg)
    if err != nil {
        return nil, nil, fmt.Errorf("create pgxpool: %w", err)
    }
    sqldb := stdlib.OpenDBFromPool(pool)
    return pool, sqldb, nil
}
