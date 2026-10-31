package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgxdriver"
	"go.opentelemetry.io/otel"
)

// Config holds database connection configuration.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DB wraps the Bun database instance with additional utilities.
type DB struct {
	bunDB  *bun.DB
	log    *slog.Logger
	tracer interface{}
}

// New creates a new database connection using Bun with pgx driver.
func New(ctx context.Context, cfg Config, log *slog.Logger) (*DB, error) {
	sqldb, err := initDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	bunDB := bun.NewDB(sqldb, pgdialect.New())

	// Configure connection pool
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection
	if err := bunDB.PingContext(ctx); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("database connected",
		"maxOpenConns", cfg.MaxOpenConns,
		"maxIdleConns", cfg.MaxIdleConns,
	)

	return &DB{
		bunDB:  bunDB,
		log:    log,
		tracer: otel.GetTracerProvider().Tracer("github.com/superindexer/indexer"),
	}, nil
}

// initDB creates a database/sql DB with pgx driver.
func initDB(ctx context.Context, cfg Config) (*sql.DB, error) {
	sqldb, err := sql.OpenContext(ctx, pgxdriver.Name, cfg.DSN)
	if err != nil {
		return nil, err
	}

	return sqldb, nil
}

// Bun returns the underlying bun.DB instance.
func (d *DB) Bun() *bun.DB {
	return d.bunDB
}

// DB returns the underlying database/sql.DB instance.
func (d *DB) DB() *sql.DB {
	return d.bunDB.DB()
}

// Tracer returns the OpenTelemetry tracer.
func (d *DB) Tracer() interface{} {
	return d.tracer
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.bunDB.DB().Close()
}

// NewTx creates a new Bun transaction.
func (d *DB) NewTx(ctx context.Context) (*bun.Tx, error) {
	return d.bunDB.NewTx()
}

// WithTx creates a new transaction and runs the given function.
// If the function returns an error, the transaction is rolled back.
// If the function returns nil, the transaction is committed.
func (d *DB) WithTx(ctx context.Context, fn func(context.Context, *bun.Tx) error) error {
	tx, err := d.bunDB.NewTx()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
