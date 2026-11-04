package migrations

import (
    "context"
    "database/sql"
    "testing"
    "time"

    _ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
    dsn := "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatalf("open db: %v", err)
    }
    // Ensure clean
    db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
    return db, func() { db.Close() }
}

func TestDriverApplyAndCurrent(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    mig := New(db)
    if err := mig.ensureSchemaTable(ctx); err != nil {
        t.Fatalf("ensure: %v", err)
    }

    // Apply migration 1
    if err := mig.Apply(ctx, 1, "CREATE TABLE foo(id INT);"); err != nil {
        t.Fatalf("apply 1: %v", err)
    }

    v, err := mig.Current(ctx)
    if err != nil || v != 1 {
        t.Fatalf("current after 1: %v, %d", err, v)
    }

    // Apply migration 2
    if err := mig.Apply(ctx, 2, "CREATE TABLE bar(id INT);"); err != nil {
        t.Fatalf("apply 2: %v", err)
    }

    v, _ = mig.Current(ctx)
    if v != 2 {
        t.Fatalf("current after 2: %d", v)
    }
}

func TestDriverRollbackOnFailure(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    mig := New(db)
    if err := mig.ensureSchemaTable(ctx); err != nil {
        t.Fatalf("ensure: %v", err)
    }

    // Apply a failing migration
    badSQL := "CREATE TABLE baz(id INT);\nINVALID SQL;"
    if err := mig.Apply(ctx, 1, badSQL); err == nil {
        t.Fatalf("expected error from bad migration")
    }

    // Ensure version not recorded
    v, err := mig.Current(ctx)
    if err != nil {
        t.Fatalf("current after failure: %v", err)
    }
    if v != 0 {
        t.Fatalf("expected version 0 after rollback, got %d", v)
    }
}
