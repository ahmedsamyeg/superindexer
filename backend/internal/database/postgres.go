package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.opentelemetry.io/otel"
)

// DB wraps a pgxpool pool with Bun ORM integration.
type DB struct {
	pool   *pgxpool.Pool
	bunDB  *bun.DB
	log    *slog.Logger
	tracer interface{}
}

// New creates a new database connection using Bun with direct pgxpool.
func New(ctx context.Context, cfg Config, log *slog.Logger) (*DB, error) {
	pool, sqldb, err := initPool(ctx, cfg)
	if err != nil {
		return nil, err
	}

	bunDB := bun.NewDB(sqldb, pgdialect.New())

	// Verify connection
	if err := bunDB.PingContext(ctx); err != nil {
		pool.Close()
		sqldb.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("database connected via pgxpool",
		"maxOpenConns", cfg.MaxOpenConns,
		"maxIdleConns", cfg.MaxIdleConns,
	)

	return &DB{
		pool:   pool,
		bunDB:  bunDB,
		log:    log,
		tracer: otel.GetTracerProvider().Tracer("github.com/superindexer/indexer"),
	}, nil
}

// initPool creates a pgxpool pool and a *sql.DB backed by it via stdlib.
func initPool(ctx context.Context, cfg Config) (*pgxpool.Pool, *sql.DB, error) {
	// Parse and configure pgxpool
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("parse DSN: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Create the pgxpool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("create pgxpool: %w", err)
	}

	// Create a *sql.DB that wraps the pgxpool for Bun compatibility
	sqldb := stdlib.OpenDBFromPool(pool)

	return pool, sqldb, nil
}

// Bun returns the underlying bun.DB instance.
func (d *DB) Bun() *bun.DB {
	return d.bunDB
}

// Pool returns the underlying pgxpool.Pool for direct pgx access.
func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

// Tracer returns the OpenTelemetry tracer.
func (d *DB) Tracer() interface{} {
	return d.tracer
}

// Close closes the database connection pool.
func (d *DB) Close() error {
	d.pool.Close()
	return nil
}

// NewTx creates a new Bun transaction.
func (d *DB) NewTx(ctx context.Context) (*bun.Tx, error) {
	tx, err := d.bunDB.BeginTx(ctx, nil)
	return &tx, err
}

// WithTx creates a new transaction and runs the given function.
// If the function returns an error, the transaction is rolled back.
// If the function returns nil, the transaction is committed.
func (d *DB) WithTx(ctx context.Context, fn func(context.Context, *bun.Tx) error) error {
	tx, err := d.bunDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(ctx, &tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// PoolStats returns the current pgxpool connection pool statistics.
func (d *DB) PoolStats() *PoolStats {
	stats := d.pool.Stat()
	return &PoolStats{
		MaxOpenConns:     int(stats.MaxConns()),
		OpenConnections:  int(stats.TotalConns()),
		InUse:            int(stats.AcquiredConns()),
		Idle:             int(stats.IdleConns()),
		WaitCount:        stats.AcquireCount(),
		WaitDuration:     stats.AcquireDuration(),
		MaxIdleClosed:    stats.MaxIdleDestroyCount(),
		MaxLifetimeClosed: stats.MaxLifetimeDestroyCount(),
	}
}

// PoolStats holds connection pool statistics.
type PoolStats struct {
	MaxOpenConns      int           `json:"max_open_connections"`
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}
