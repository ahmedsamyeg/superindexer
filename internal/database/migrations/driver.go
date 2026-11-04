package migrations

import (
    "context"
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

// DB is a thin wrapper around sql.DB that knows how to apply migrations.
type DB struct {
    *sql.DB
}

// New creates a new migration-enabled database.
func New(db *sql.DB) *DB {
    return &DB{db}
}

// ensureSchemaTable makes sure the schema_migrations table exists.
func (m *DB) ensureSchemaTable(ctx context.Context) error {
    const create = `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version BIGINT PRIMARY KEY,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
        );`
    _, err := m.ExecContext(ctx, create)
    return err
}

// Current returns the current schema version.
func (m *DB) Current(ctx context.Context) (int64, error) {
    var v int64
    err := m.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&v)
    return v, err
}

// Apply applies a single migration if needed.
func (m *DB) Apply(ctx context.Context, version int64, up string) error {
    // Check if already applied
    cur, err := m.Current(ctx)
    if err != nil {
        return err
    }
    if cur >= version {
        return nil
    }
    tx, err := m.BeginTx(ctx, &sql.TxOptions{})
    if err != nil {
        return err
    }
    _, err = tx.ExecContext(ctx, up)
    if err != nil {
        tx.Rollback()
        return err
    }
    _, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", version)
    if err != nil {
        tx.Rollback()
        return err
    }
    return tx.Commit()
}
