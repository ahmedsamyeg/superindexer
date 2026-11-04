package migrations

import (
    "context"
    "database/sql"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
)

// Runner applies all pending migrations in the migrations directory.
// It expects files named like "001_initial.sql", "002_add_posts.sql" etc.
// The numeric prefix determines order. The file is parsed for an
// "-- +goose Up" block; everything after that line until a
// "-- +goose Down" or EOF is considered the up migration.
// The driver handles applying each migration.

type Runner struct {
    // Path to the directory containing .sql migration files
    Dir string
}

// NewRunner creates a new Runner for the given directory.
func NewRunner(dir string) *Runner {
    return &Runner{Dir: dir}
}

// Run applies all pending migrations using the provided *sql.DB.
func (r *Runner) Run(ctx context.Context, db *sql.DB) error {
    // Load all .sql files
    entries, err := fs.ReadDir(os.DirFS(r.Dir), ".")
    if err != nil {
        return fmt.Errorf("reading migrations dir %s: %w", r.Dir, err)
    }

    var files []string
    for _, e := range entries {
        if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
            files = append(files, e.Name())
        }
    }

    // Sort by numeric prefix
    sort.Slice(files, func(i, j int) bool {
        vi := migrationVersion(files[i])
        vj := migrationVersion(files[j])
        return vi < vj
    })

    migDB := New(db)
    if err := migDB.ensureSchemaTable(ctx); err != nil {
        return fmt.Errorf("ensure schema table: %w", err)
    }

    for _, f := range files {
        path := filepath.Join(r.Dir, f)
        contentBytes, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("reading migration %s: %w", f, err)
        }
        upSQL := parseUp(string(contentBytes))
        if strings.TrimSpace(upSQL) == "" {
            continue // skip if no up block
        }
        version := migrationVersion(f)
        if err := migDB.Apply(ctx, int64(version), upSQL); err != nil {
            return fmt.Errorf("applying migration %s: %w", f, err)
        }
    }
    return nil
}

// migrationVersion extracts the leading numeric prefix from a filename.
func migrationVersion(name string) int {
    parts := strings.SplitN(name, "_", 2)
    if len(parts) == 0 {
        return 0
    }
    v, err := strconv.Atoi(parts[0])
    if err != nil {
        return 0
    }
    return v
}

// parseUp extracts the SQL between "-- +goose Up" and "-- +goose Down".
func parseUp(src string) string {
    var sb strings.Builder
    lines := strings.Split(src, "\n")
    inUp := false
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        if trimmed == "-- +goose Up" {
            inUp = true
            continue
        }
        if trimmed == "-- +goose Down" {
            break
        }
        if inUp {
            sb.WriteString(line)
            sb.WriteByte('\n')
        }
    }
    return sb.String()
}
