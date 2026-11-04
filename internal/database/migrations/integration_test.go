package migrations

import (
    "context"
    "database/sql"
    "testing"
    "time"

    _ "github.com/lib/pq"
)

// This integration test requires a running Postgres instance at localhost:5432
// with database "testdb" and user/password postgres. It will create the schema
// migrations table, apply a simple migration and verify the version.
func TestIntegrationApply(t *testing.T) {
    dsn := "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Skip("could not connect to postgres, skipping integration test")
    }
    defer db.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mig := New(db)
    if err := mig.ensureSchemaTable(ctx); err != nil {
        t.Fatalf("ensure: %v", err)
    }

    if err := mig.Apply(ctx, 1, "CREATE TABLE qux(id INT);"); err != nil {
        t.Fatalf("apply: %v", err)
    }

    v, err := mig.Current(ctx)
    if err != nil || v != 1 {
        t.Fatalf("current: %v, %d", err, v)
    }
}
