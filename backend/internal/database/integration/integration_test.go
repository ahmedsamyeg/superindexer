package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/superindexer/indexer/internal/database"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/observability"
	"github.com/uptrace/bun"
)

// getTestDSN returns the PostgreSQL DSN for integration tests.
// It reads from TEST_DATABASE_DSN environment variable, or falls back to localhost.
func getTestDSN() string {
	if dsn := os.Getenv("TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://indexer:indexer-dev@localhost:5432/indexer_test?sslmode=disable"
}



// createTestDB creates a new database connection for testing.
func createTestDB(t *testing.T) *database.DB {
	t.Helper()

	cfg := database.Config{
		DSN:             getTestDSN(),
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	log := observability.NewLogger("debug", "json")

	db, err := database.New(context.Background(), cfg, log)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return db
}

// setupTestDB runs migrations and prepares the database for testing.
func setupTestDB(t *testing.T, db *database.DB) {
	t.Helper()

	// Get the migration file path
	migDir := filepath.Join("..", "..", "..", "internal", "database", "migrations")
	absMigDir, err := filepath.Abs(migDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for migrations: %v", err)
	}

	migrationFile := filepath.Join(absMigDir, "001_initial.up.sql")

	// Copy migration file into the Docker container
	copyCmd := exec.Command("docker", "cp", migrationFile, "postgres-17:/tmp/001_initial.sql")
	if err := copyCmd.Run(); err != nil {
		t.Fatalf("failed to copy migration file: %v", err)
	}

	// Run psql inside the container
	output, err := exec.Command("docker", "exec", "postgres-17", "psql", "-U", "indexer", "-d", "indexer_test", "-f", "/tmp/001_initial.sql").CombinedOutput()
	if err != nil {
		t.Logf("Migration output: %s", string(output))
		t.Fatalf("failed to run migration: %v", err)
	}
	t.Logf("Migration applied successfully")
}

// teardownTestDB drops all tables and closes the database connection.
func teardownTestDB(t *testing.T, db *database.DB) {
	t.Helper()

	// Drop all tables
	_, err := db.Bun().DB.ExecContext(context.Background(), `
		DROP TABLE IF EXISTS system_settings CASCADE;
		DROP TABLE IF EXISTS bookmarks CASCADE;
		DROP TABLE IF EXISTS user_search_history CASCADE;
		DROP TABLE IF EXISTS indexing_checkpoints CASCADE;
		DROP TABLE IF EXISTS indexing_jobs CASCADE;
		DROP TABLE IF EXISTS spotnet_posts CASCADE;
		DROP TABLE IF EXISTS spotnet_sources CASCADE;
		DROP TABLE IF EXISTS articles CASCADE;
		DROP TABLE IF EXISTS categories CASCADE;
		DROP TABLE IF EXISTS newsgroups CASCADE;
		DROP TABLE IF EXISTS nntp_servers CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
	`)
	if err != nil {
		t.Logf("warning: failed to drop tables: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Logf("warning: failed to close database: %v", err)
	}
}

// TestDatabaseConnection verifies that the pgxpool connection works correctly.
func TestDatabaseConnection(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	// Verify the pool is accessible
	pool := db.Pool()
	if pool == nil {
		t.Fatal("pool should not be nil")
	}

	// Verify Bun DB is accessible
	bunDB := db.Bun()
	if bunDB == nil {
		t.Fatal("bunDB should not be nil")
	}

	// Verify we can ping the database
	if err := bunDB.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify pool stats are accessible
	stats := pool.Stat()
	if stats == nil {
		t.Fatal("pool stats should not be nil")
	}
}

// TestMigrationExecution verifies that migrations run correctly.
func TestMigrationExecution(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	// Verify tables exist by querying them
	var count int
	err := db.Bun().DB.QueryRowContext(context.Background(), "SELECT count(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("users table should exist: %v", err)
	}

	// Verify newsgroups table exists
	err = db.Bun().DB.QueryRowContext(context.Background(), "SELECT count(*) FROM newsgroups").Scan(&count)
	if err != nil {
		t.Fatalf("newsgroups table should exist: %v", err)
	}

	// Verify articles table exists
	err = db.Bun().DB.QueryRowContext(context.Background(), "SELECT count(*) FROM articles").Scan(&count)
	if err != nil {
		t.Fatalf("articles table should exist: %v", err)
	}

	// Verify spotnet_sources table exists
	err = db.Bun().DB.QueryRowContext(context.Background(), "SELECT count(*) FROM spotnet_sources").Scan(&count)
	if err != nil {
		t.Fatalf("spotnet_sources table should exist: %v", err)
	}

	// Verify spotnet_posts table exists
	err = db.Bun().DB.QueryRowContext(context.Background(), "SELECT count(*) FROM spotnet_posts").Scan(&count)
	if err != nil {
		t.Fatalf("spotnet_posts table should exist: %v", err)
	}

	// Verify indexing_jobs table exists
	err = db.Bun().DB.QueryRowContext(context.Background(), "SELECT count(*) FROM indexing_jobs").Scan(&count)
	if err != nil {
		t.Fatalf("indexing_jobs table should exist: %v", err)
	}

	// Verify FTS extensions are enabled
	var extCount int
	err = db.Bun().DB.QueryRowContext(context.Background(),
		"SELECT count(*) FROM pg_extension WHERE extname IN ('unaccent', 'pg_trgm')").Scan(&extCount)
	if err != nil {
		t.Fatalf("extensions check failed: %v", err)
	}
	if extCount != 2 {
		t.Fatalf("expected 2 extensions, got %d", extCount)
	}

	// Verify GIN indexes exist
	var idxCount int
	err = db.Bun().DB.QueryRowContext(context.Background(),
		"SELECT count(*) FROM pg_indexes WHERE indexname IN ('idx_articles_search_vector', 'idx_spotnet_posts_search_vector')").Scan(&idxCount)
	if err != nil {
		t.Fatalf("index check failed: %v", err)
	}
	if idxCount != 2 {
		t.Fatalf("expected 2 GIN indexes, got %d", idxCount)
	}
}

// TestUserRepositoryCRUD tests user repository operations.
func TestUserRepositoryCRUD(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	repo := repositories.NewUserRepository(db.Bun())

	// Create a user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$hash",
		Role:         "user",
		IsActive:     true,
	}

	ctx := context.Background()
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("user ID should be set after create")
	}

	// Get user by ID
	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to get user by ID: %v", err)
	}
	if got.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got '%s'", got.Username)
	}

	// Get user by username
	got, err = repo.GetByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("failed to get user by username: %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, got.ID)
	}

	// Get user by email
	got, err = repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, got.ID)
	}

	// Update user
	got.Email = "updated@example.com"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	// Verify update
	got, err = repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to get updated user: %v", err)
	}
	if got.Email != "updated@example.com" {
		t.Fatalf("expected email 'updated@example.com', got '%s'", got.Email)
	}

	// Count users
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 user, got %d", count)
	}

	// List users with pagination
	result, err := repo.List(ctx, repositories.Pagination{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	// Update password hash
	if err := repo.UpdatePasswordHash(ctx, user.ID, "$argon2id$v=19$m=65536,t=3,p=4$newhash"); err != nil {
		t.Fatalf("failed to update password hash: %v", err)
	}

	// Delete user
	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	// Verify deletion
	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count users after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 users after delete, got %d", count)
	}
}

// TestTransactionHandling tests transaction operations.
func TestTransactionHandling(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	// Create a user in a transaction
	err := db.WithTx(context.Background(), func(ctx context.Context, tx *bun.Tx) error {
		user := &models.User{
			Username:     "txuser",
			Email:        "tx@example.com",
			PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$hash",
			Role:         "user",
			IsActive:     true,
		}

		_, err := tx.NewInsert().Model(user).Exec(ctx)
		if err != nil {
			return err
		}

		// Verify the user was inserted
		var count int
		err = tx.NewSelect().Model((*models.User)(nil)).
			Where("username = ?", "txuser").
			Column("count(*)").
			Scan(ctx, &count)
		if err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("expected 1 user in transaction, got %d", count)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Verify the user was committed
	repo := repositories.NewUserRepository(db.Bun())
	got, err := repo.GetByUsername(context.Background(), "txuser")
	if err != nil {
		t.Fatalf("failed to get user after transaction: %v", err)
	}
	if got.Username != "txuser" {
		t.Fatalf("expected username 'txuser', got '%s'", got.Username)
	}
}

// TestTransactionRollback tests that failed transactions are rolled back.
func TestTransactionRollback(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	// Try to create a user in a transaction that fails
	err := db.WithTx(context.Background(), func(ctx context.Context, tx *bun.Tx) error {
		user := &models.User{
			Username:     "rollbackuser",
			Email:        "rollback@example.com",
			PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$hash",
			Role:         "user",
			IsActive:     true,
		}

		_, err := tx.NewInsert().Model(user).Exec(ctx)
		if err != nil {
			return err
		}

		// Force a rollback by returning an error
		return fmt.Errorf("forced rollback")
	})
	if err == nil {
		t.Fatal("expected transaction to fail")
	}

	// Verify the user was not committed
	repo := repositories.NewUserRepository(db.Bun())
	_, err = repo.GetByUsername(context.Background(), "rollbackuser")
	if err == nil {
		t.Fatal("expected user not to exist after rollback")
	}
}

// TestNewsgroupRepository tests newsgroup repository operations.
func TestNewsgroupRepository(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	repo := repositories.NewNewsgroupRepository(db.Bun())

	// Create a newsgroup
	group := &models.Newsgroup{
		Name:         "test.newsgroup",
		Description:  "Test newsgroup",
		Enabled:      true,
		IndexingPriority: 5,
	}

	ctx := context.Background()
	if err := repo.Create(ctx, group); err != nil {
		t.Fatalf("failed to create newsgroup: %v", err)
	}

	if group.ID == 0 {
		t.Fatal("newsgroup ID should be set after create")
	}

	// Get by ID
	got, err := repo.GetByID(ctx, group.ID)
	if err != nil {
		t.Fatalf("failed to get newsgroup by ID: %v", err)
	}
	if got.Name != "test.newsgroup" {
		t.Fatalf("expected name 'test.newsgroup', got '%s'", got.Name)
	}

	// Get by name
	got, err = repo.GetByName(ctx, "test.newsgroup")
	if err != nil {
		t.Fatalf("failed to get newsgroup by name: %v", err)
	}
	if got.ID != group.ID {
		t.Fatalf("expected ID %d, got %d", group.ID, got.ID)
	}

	// Get all
	groups, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all newsgroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 newsgroup, got %d", len(groups))
	}

	// Get enabled
	groups, err = repo.GetEnabled(ctx)
	if err != nil {
		t.Fatalf("failed to get enabled newsgroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 enabled newsgroup, got %d", len(groups))
	}

	// Update article range
	if err := repo.UpdateArticleRange(ctx, group.ID, 1000, 2000); err != nil {
		t.Fatalf("failed to update article range: %v", err)
	}

	// Update last indexed
	now := time.Now().UTC()
	if err := repo.UpdateLastIndexed(ctx, group.ID, 1500, &now); err != nil {
		t.Fatalf("failed to update last indexed: %v", err)
	}

	// Count
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count newsgroups: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 newsgroup, got %d", count)
	}

	// Count enabled
	count, err = repo.CountEnabled(ctx)
	if err != nil {
		t.Fatalf("failed to count enabled newsgroups: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 enabled newsgroup, got %d", count)
	}

	// Delete
	if err := repo.Delete(ctx, group.ID); err != nil {
		t.Fatalf("failed to delete newsgroup: %v", err)
	}

	// Verify deletion
	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 newsgroups after delete, got %d", count)
	}
}

// TestArticleRepository tests article repository operations.
func TestArticleRepository(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	// Create a newsgroup first
	ngRepo := repositories.NewNewsgroupRepository(db.Bun())
	ng := &models.Newsgroup{
		Name:         "test.articles",
		Description:  "Test newsgroup for articles",
		Enabled:      true,
		IndexingPriority: 5,
	}
	if err := ngRepo.Create(context.Background(), ng); err != nil {
		t.Fatalf("failed to create newsgroup: %v", err)
	}

	repo := repositories.NewArticleRepository(db.Bun())

	// Create an article
	article := &models.Article{
		MessageID:   "<test-article-1@example.com>",
		NewsgroupID: ng.ID,
		Subject:     "Test Article Subject",
		Author:      "Test Author",
		PostedAt:    time.Now().UTC(),
		ArticleSize: 1024,
	}

	ctx := context.Background()
	if err := repo.Create(ctx, article); err != nil {
		t.Fatalf("failed to create article: %v", err)
	}

	if article.ID == 0 {
		t.Fatal("article ID should be set after create")
	}

	// Get by ID
	got, err := repo.GetByID(ctx, article.ID)
	if err != nil {
		t.Fatalf("failed to get article by ID: %v", err)
	}
	if got.MessageID != "<test-article-1@example.com>" {
		t.Fatalf("expected message_id '<test-article-1@example.com>', got '%s'", got.MessageID)
	}

	// Get by message ID
	got, err = repo.GetByMessageID(ctx, "<test-article-1@example.com>")
	if err != nil {
		t.Fatalf("failed to get article by message ID: %v", err)
	}
	if got.ID != article.ID {
		t.Fatalf("expected ID %d, got %d", article.ID, got.ID)
	}

	// Check exists
	exists, err := repo.ExistsByMessageID(ctx, "<test-article-1@example.com>")
	if err != nil {
		t.Fatalf("failed to check exists: %v", err)
	}
	if !exists {
		t.Fatal("article should exist")
	}

	// Count
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count articles: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 article, got %d", count)
	}

	// Count by newsgroup
	count, err = repo.CountByNewsgroup(ctx, ng.ID)
	if err != nil {
		t.Fatalf("failed to count by newsgroup: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 article in newsgroup, got %d", count)
	}

	// Update
	got.Subject = "Updated Subject"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("failed to update article: %v", err)
	}

	// Verify update
	got, err = repo.GetByID(ctx, article.ID)
	if err != nil {
		t.Fatalf("failed to get updated article: %v", err)
	}
	if got.Subject != "Updated Subject" {
		t.Fatalf("expected subject 'Updated Subject', got '%s'", got.Subject)
	}

	// List by newsgroup with pagination
	result, err := repo.ListByNewsgroup(ctx, ng.ID, repositories.Pagination{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("failed to list articles: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	// Delete
	if err := repo.Delete(ctx, article.ID); err != nil {
		t.Fatalf("failed to delete article: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, article.ID)
	if err == nil {
		t.Fatal("expected error getting deleted article")
	}
}

// TestArticleBulkInsert tests bulk article insertion.
func TestArticleBulkInsert(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	// Create a newsgroup first
	ngRepo := repositories.NewNewsgroupRepository(db.Bun())
	ng := &models.Newsgroup{
		Name:         "test.bulk",
		Description:  "Test newsgroup for bulk insert",
		Enabled:      true,
		IndexingPriority: 5,
	}
	if err := ngRepo.Create(context.Background(), ng); err != nil {
		t.Fatalf("failed to create newsgroup: %v", err)
	}

	repo := repositories.NewArticleRepository(db.Bun())

	// Create bulk articles
	articles := make([]*models.Article, 10)
	for i := 0; i < 10; i++ {
		articles[i] = &models.Article{
			MessageID:   fmt.Sprintf("<bulk-article-%d@example.com>", i),
			NewsgroupID: ng.ID,
			Subject:     fmt.Sprintf("Bulk Article %d", i),
			Author:      "Bulk Author",
			PostedAt:    time.Now().UTC(),
			ArticleSize: 1024 + int64(i)*100,
		}
	}

	ctx := context.Background()
	if err := repo.CreateBulk(ctx, articles); err != nil {
		t.Fatalf("failed to bulk insert articles: %v", err)
	}

	// Verify count
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count articles: %v", err)
	}
	if count != 10 {
		t.Fatalf("expected 10 articles, got %d", count)
	}

	// Test idempotency - insert same articles again
	if err := repo.CreateBulk(ctx, articles); err != nil {
		t.Fatalf("failed to bulk insert duplicates: %v", err)
	}

	// Count should still be 10 (no duplicates)
	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count articles after duplicate insert: %v", err)
	}
	if count != 10 {
		t.Fatalf("expected 10 articles after duplicate insert, got %d", count)
	}
}

// TestSearchRepository tests search functionality.
func TestSearchRepository(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	// Create a newsgroup first
	ngRepo := repositories.NewNewsgroupRepository(db.Bun())
	ng := &models.Newsgroup{
		Name:         "test.search",
		Description:  "Test newsgroup for search",
		Enabled:      true,
		IndexingPriority: 5,
	}
	if err := ngRepo.Create(context.Background(), ng); err != nil {
		t.Fatalf("failed to create newsgroup: %v", err)
	}

	// Insert test articles
	articleRepo := repositories.NewArticleRepository(db.Bun())
	articles := []*models.Article{
		{
			MessageID:   "<search-test-1@example.com>",
			NewsgroupID: ng.ID,
			Subject:     "Go Programming Language",
			Author:      "Go Developer",
			PostedAt:    time.Now().UTC(),
			ArticleSize: 2048,
		},
		{
			MessageID:   "<search-test-2@example.com>",
			NewsgroupID: ng.ID,
			Subject:     "Python Data Analysis",
			Author:      "Python Developer",
			PostedAt:    time.Now().UTC(),
			ArticleSize: 4096,
		},
		{
			MessageID:   "<search-test-3@example.com>",
			NewsgroupID: ng.ID,
			Subject:     "Rust Systems Programming",
			Author:      "Rust Developer",
			PostedAt:    time.Now().UTC(),
			ArticleSize: 1024,
		},
	}

	ctx := context.Background()
	for _, a := range articles {
		if err := articleRepo.Create(ctx, a); err != nil {
			t.Fatalf("failed to create article: %v", err)
		}
	}

	searchRepo := repositories.NewSearchRepository(db.Bun())

	// Test search by query
	result, err := searchRepo.SearchArticles(ctx, models.SearchQuery{
		Query:   "programming",
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("failed to search articles: %v", err)
	}

	// Should find at least 2 articles (Go and Rust both contain "programming" in subject)
	if result.Total < 2 {
		t.Fatalf("expected at least 2 results for 'programming', got %d", result.Total)
	}

	// Test search with author filter
	result, err = searchRepo.SearchArticles(ctx, models.SearchQuery{
		Query:   "Go",
		Author:  "Go Developer",
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("failed to search with author filter: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 result for 'Go' by 'Go Developer', got %d", result.Total)
	}

	// Test search with size filter
	result, err = searchRepo.SearchArticles(ctx, models.SearchQuery{
		Query:   "programming",
		MinSize: intPtr(2000),
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("failed to search with size filter: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 results with min size 2000, got %d", result.Total)
	}

	// Test facets
	facets, err := searchRepo.GetFacets(ctx, models.SearchQuery{
		Query:   "programming",
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("failed to get facets: %v", err)
	}
	if len(facets) == 0 {
		t.Fatal("expected non-empty facets")
	}
}

// TestCategoryRepository tests category repository operations.
func TestCategoryRepository(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	repo := repositories.NewCategoryRepository(db.Bun())

	ctx := context.Background()

	// Create parent category
	parent := &models.Category{
		Name: "Technology",
		Slug: "technology",
	}
	if err := repo.Create(ctx, parent); err != nil {
		t.Fatalf("failed to create parent category: %v", err)
	}

	// Create child category
	child := &models.Category{
		Name:     "Programming",
		Slug:     "programming",
		ParentID: &parent.ID,
	}
	if err := repo.Create(ctx, child); err != nil {
		t.Fatalf("failed to create child category: %v", err)
	}

	// Get by ID
	got, err := repo.GetByID(ctx, parent.ID)
	if err != nil {
		t.Fatalf("failed to get category by ID: %v", err)
	}
	if got.Name != "Technology" {
		t.Fatalf("expected name 'Technology', got '%s'", got.Name)
	}

	// Get by slug
	got, err = repo.GetBySlug(ctx, "programming")
	if err != nil {
		t.Fatalf("failed to get category by slug: %v", err)
	}
	if got.Name != "Programming" {
		t.Fatalf("expected name 'Programming', got '%s'", got.Name)
	}

	// Get by name
	got, err = repo.GetByName(ctx, "Technology")
	if err != nil {
		t.Fatalf("failed to get category by name: %v", err)
	}
	if got.ID != parent.ID {
		t.Fatalf("expected ID %d, got %d", parent.ID, got.ID)
	}

	// Get all
	categories, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all categories: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}

	// Get children
	children, err := repo.GetChildren(ctx, parent.ID)
	if err != nil {
		t.Fatalf("failed to get children: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0].Name != "Programming" {
		t.Fatalf("expected child name 'Programming', got '%s'", children[0].Name)
	}

	// Delete parent (should cascade to child or fail gracefully)
	if err := repo.Delete(ctx, parent.ID); err != nil {
		t.Fatalf("failed to delete parent category: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, parent.ID)
	if err == nil {
		t.Fatal("expected error getting deleted category")
	}
}

// TestIndexingJobRepository tests indexing job repository operations.
func TestIndexingJobRepository(t *testing.T) {
	db := createTestDB(t)
	defer teardownTestDB(t, db)

	setupTestDB(t, db)

	repo := repositories.NewIndexingJobRepository(db.Bun())

	ctx := context.Background()

	// Create a job
	job := &models.IndexingJob{
		JobType:    "nntp_sync",
		Status:     "pending",
		Priority:   5,
		Progress:   0,
		SourceID:   intPtr(1),
		NewsgroupID: intPtr(1),
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	if job.ID == 0 {
		t.Fatal("job ID should be set after create")
	}

	// Get by ID
	got, err := repo.GetByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("failed to get job by ID: %v", err)
	}
	if got.JobType != "nntp_sync" {
		t.Fatalf("expected job type 'nntp_sync', got '%s'", got.JobType)
	}

	// Get pending jobs
	jobs, err := repo.GetPendingJobs(ctx)
	if err != nil {
		t.Fatalf("failed to get pending jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 pending job, got %d", len(jobs))
	}

	// Update status to running
	if err := repo.UpdateStatus(ctx, job.ID, "running"); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	// Update progress
	if err := repo.UpdateProgress(ctx, job.ID, 0.5, 500, 10); err != nil {
		t.Fatalf("failed to update progress: %v", err)
	}

	// Verify progress
	got, err = repo.GetByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("failed to get job after progress update: %v", err)
	}
	if got.Progress != 0.5 {
		t.Fatalf("expected progress 0.5, got %f", got.Progress)
	}
	if got.RecordsProcessed != 500 {
		t.Fatalf("expected records_processed 500, got %d", got.RecordsProcessed)
	}

	// Update error
	if err := repo.UpdateError(ctx, job.ID, "connection timeout"); err != nil {
		t.Fatalf("failed to update error: %v", err)
	}

	// Complete the job
	if err := repo.Complete(ctx, job.ID); err != nil {
		t.Fatalf("failed to complete job: %v", err)
	}

	// Get by status
	jobs, err = repo.GetByStatus(ctx, "completed")
	if err != nil {
		t.Fatalf("failed to get completed jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 completed job, got %d", len(jobs))
	}

	// Count by status
	count, err := repo.CountByStatus(ctx, "completed")
	if err != nil {
		t.Fatalf("failed to count completed jobs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 completed job, got %d", count)
	}

	// Count active
	count, err = repo.CountActive(ctx)
	if err != nil {
		t.Fatalf("failed to count active jobs: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 active jobs, got %d", count)
	}

	// Delete
	if err := repo.Delete(ctx, job.ID); err != nil {
		t.Fatalf("failed to delete job: %v", err)
	}
}

// Helper functions
func intPtr(i int64) *int64 {
	return &i
}
