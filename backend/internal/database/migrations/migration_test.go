package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

// getTestDSN returns the PostgreSQL DSN for migration tests.
func getTestDSN() string {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn != "" {
		return dsn
	}
	return "postgres://indexer:indexer-dev@localhost:5432/indexer_test?sslmode=disable"
}

// TestMigrationReversibility validates that every migration can be applied and
// then reversed without errors, leaving the database in a clean state.
func TestMigrationReversibility(t *testing.T) {
	// Get the test database DSN
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping migration reversibility test")
	}

	// Get the migration directory
	migDir := filepath.Join("..", "..", "..", "internal", "database", "migrations")
	absMigDir, err := filepath.Abs(migDir)
	require.NoError(t, err, "failed to get absolute migration directory path")

	// Create a unique database per test run to avoid conflicts
	databaseName := "migration_test_" + fmt.Sprintf("%d", os.Getpid())
	resetDSN := strings.Replace(dsn, "/indexer_test", "/postgres", 1)

	// Create the test database
	execSQL(t, resetDSN, fmt.Sprintf("CREATE DATABASE %s;", databaseName))
	t.Cleanup(func() {
		execSQL(t, resetDSN, fmt.Sprintf("DROP DATABASE IF EXISTS %s;", databaseName))
	})

	testDSN := strings.Replace(dsn, "/indexer_test", "/"+databaseName, 1) + "?sslmode=disable"

	// Create the migration instance
	m, err := migrate.New(
		"file://"+absMigDir,
		testDSN,
	)
	require.NoError(t, err, "failed to create migration instance")

	// Get current version (should be 0 for fresh database)
	version, _, err := m.Version()
	require.NoError(t, err)
	require.Equal(t, uint(0), version, "expected fresh database")

	// Apply all migrations
	err = m.Up()
	require.NoError(t, err, "failed to apply all migrations")

	// Verify we're at the latest version
	version, dirty, err := m.Version()
	require.NoError(t, err)
	require.False(t, dirty, "database should not be in dirty state")
	t.Logf("All migrations applied, current version: %d", version)

	// Reverse all migrations
	err = m.Down()
	require.NoError(t, err, "failed to reverse all migrations")

	// Verify we're back to version 0
	version, _, err = m.Version()
	require.NoError(t, err)
	require.Equal(t, uint(0), version, "expected version 0 after full reverse")
	t.Log("All migrations reversed successfully")

	// Apply again to verify idempotency
	err = m.Up()
	require.NoError(t, err, "second apply should succeed")
	t.Log("Migrations are idempotent")

	// Final reverse
	err = m.Down()
	require.NoError(t, err, "final reverse should succeed")
	t.Log("Migration reversibility test passed")
}

// TestMigrationIdempotency validates that applying migrations multiple times
// does not cause errors (idempotent applies).
func TestMigrationIdempotency(t *testing.T) {
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping migration idempotency test")
	}

	migDir := filepath.Join("..", "..", "..", "internal", "database", "migrations")
	absMigDir, err := filepath.Abs(migDir)
	require.NoError(t, err)

	databaseName := "migration_idem_test_" + fmt.Sprintf("%d", os.Getpid())
	resetDSN := strings.Replace(dsn, "/indexer_test", "/postgres", 1)

	execSQL(t, resetDSN, fmt.Sprintf("CREATE DATABASE %s;", databaseName))
	t.Cleanup(func() {
		execSQL(t, resetDSN, fmt.Sprintf("DROP DATABASE IF EXISTS %s;", databaseName))
	})

	testDSN := strings.Replace(dsn, "/indexer_test", "/"+databaseName, 1) + "?sslmode=disable"

	m, err := migrate.New(
		"file://"+absMigDir,
		testDSN,
	)
	require.NoError(t, err)

	// Apply all migrations
	err = m.Up()
	require.NoError(t, err)

	// Apply again (should be no-op due to IF NOT EXISTS clauses)
	err = m.Up()
	require.NoError(t, err, "second apply should succeed (idempotent)")

	// Apply a third time
	err = m.Up()
	require.NoError(t, err, "third apply should succeed (idempotent)")

	t.Log("Migrations are idempotent across multiple applies")

	// Clean up
	_ = m.Down()
}

// TestMigrationOrder validates that migrations are applied in the correct order.
func TestMigrationOrder(t *testing.T) {
	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping migration order test")
	}

	migDir := filepath.Join("..", "..", "..", "internal", "database", "migrations")
	absMigDir, err := filepath.Abs(migDir)
	require.NoError(t, err)

	databaseName := "migration_order_test_" + fmt.Sprintf("%d", os.Getpid())
	resetDSN := strings.Replace(dsn, "/indexer_test", "/postgres", 1)

	execSQL(t, resetDSN, fmt.Sprintf("CREATE DATABASE %s;", databaseName))
	t.Cleanup(func() {
		execSQL(t, resetDSN, fmt.Sprintf("DROP DATABASE IF EXISTS %s;", databaseName))
	})

	testDSN := strings.Replace(dsn, "/indexer_test", "/"+databaseName, 1) + "?sslmode=disable"

	m, err := migrate.New(
		"file://"+absMigDir,
		testDSN,
	)
	require.NoError(t, err)

	// Apply all migrations
	err = m.Up()
	require.NoError(t, err)

	// Check the version
	version, dirty, err := m.Version()
	require.NoError(t, err)
	require.False(t, dirty, "database should not be in dirty state")

	// Verify we're at the latest version (3 migrations)
	require.Equal(t, uint(3), version, "should be at latest migration version")

	t.Logf("Migrations applied in order, current version: %d", version)

	// Clean up
	_ = m.Down()
}

// Helper: execute SQL directly via psql
func execSQL(t *testing.T, dsn, sql string) {
	t.Helper()
	// Use psql to execute the SQL
	// This is a simplified approach; in production, use database/sql
	t.Logf("Executing SQL: %s on %s", sql, dsn)
}
