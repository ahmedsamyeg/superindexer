# Database Migrations

This directory contains all database migrations for the Usenet + Spotnet Indexer.

## Migration Structure

Migrations follow the pattern:
- `000001_init.down.sql` - Down migration (reverses the up migration)
- `000001_init.up.sql` - Up migration (creates tables and schema)

## Running Migrations

### Using the migrate command:

```bash
# Run all pending migrations (up)
migrate -database "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable" -path ./backend/internal/database/migrations up

# Run specific migration (e.g., to version 2)
migrate -database "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable" -path ./backend/internal/database/migrations up 2

# Rollback to previous version (e.g., from version 2 to 1)
migrate -database "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable" -path ./backend/internal/database/migrations down 1

# Rollback all migrations
migrate -database "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable" -path ./backend/internal/database/migrations down
```

### Using Docker:

```bash
# Run migrations in Docker
docker compose -f deployments/docker/docker-compose.yml run --rm migrate up

# Or with specific database URL
docker run --rm -v $(pwd)/backend/internal/database/migrations:/migrations \
  migrate/migrate -database "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable" \
  -path /migrations up
```

## Migration Best Practices

1. **Versioned Migrations**: Each migration has a unique version number
2. **Reversible Changes**: Every up migration should have a corresponding down migration  
3. **Safe Operations**: Use `IF NOT EXISTS` and proper constraints to avoid conflicts
4. **Testing**: Test migrations on a copy of production data when possible

## Migration Example Format

```sql
-- +goose Up
-- SQL to apply when migrating up
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
-- SQL to apply when migrating down
DROP TABLE IF EXISTS users;
```

## Migration Order

1. `000001_init.up.sql` - Initial database schema creation
2. `000002_add_indexes.up.sql` - Add database indexes for performance
3. `000003_add_constraints.up.sql` - Add foreign key and other constraints

## Database Schema Overview

The database schema includes the following main tables:

1. **users** - User accounts and authentication
2. **nntp_servers** - Configured NNTP server connections  
3. **newsgroups** - Usenet newsgroups with article ranges
4. **articles** - Indexed Usenet articles with search vectors
5. **spotnet_sources** - Configured Spotnet source connections
6. **spotnet_posts** - Indexed Spotnet posts with metadata
7. **categories** - Content categories for organization
8. **indexing_jobs** - Background indexing job tracking

## Migration Testing

To test migrations:

1. Create a test database:
```bash
createdb indexer_test
```

2. Run migrations against the test database:
```bash
migrate -database "postgres://indexer:indexer@localhost:5432/indexer_test?sslmode=disable" -path ./backend/internal/database/migrations up
```

3. Verify the schema and data integrity.

## Troubleshooting

### Migration Conflicts
If migrations fail due to conflicts, check:
- That the database is accessible and credentials are correct
- That migration files are properly formatted
- That there are no syntax errors in the SQL

### Rollback Issues  
If rollback fails:
- Check that all down migrations are properly implemented
- Ensure the database connection is stable during rollback
- Manually verify data integrity after rollback operations