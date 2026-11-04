package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestPostgresContainer creates a PostgreSQL container for integration tests.
// It handles lifecycle: creation, migration, cleanup.
type TestPostgresContainer struct {
	Container testcontainers.Container
	DSN       string
	DBName    string
}

// StartPostgresContainer creates and starts a PostgreSQL container for testing.
func StartPostgresContainer(ctx context.Context, t *testing.T) *TestPostgresContainer {
	t.Helper()

	dbName := "indexer_test"
	dbUser := "indexer"
	dbPassword := "indexer-dev"

	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "failed to create postgres container")

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	endpoint, err := container.Endpoint(context.Background(), "")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, endpoint, dbName)

	return &TestPostgresContainer{
		Container: container,
		DSN:       dsn,
		DBName:    dbName,
	}
}

// RunMigrations applies all migrations to the test container's database.
func (tc *TestPostgresContainer) RunMigrations(ctx context.Context, t *testing.T) {
	t.Helper()

	migDir := filepath.Join("..", "..", "..", "internal", "database", "migrations")
	absMigDir, err := filepath.Abs(migDir)
	require.NoError(t, err)

	m, err := migrate.New(
		"file://"+absMigDir,
		tc.DSN,
	)
	require.NoError(t, err, "failed to create migration instance")

	err = m.Up()
	require.NoError(t, err, "failed to apply migrations")

	t.Log("Migrations applied successfully")
}
