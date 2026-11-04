-- +migrate Up

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- NNTP servers table
CREATE TABLE IF NOT EXISTS nntp_servers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 119,
    username VARCHAR(255),
    encrypted_password TEXT,
    tls_enabled BOOLEAN NOT NULL DEFAULT false,
    max_connections INTEGER NOT NULL DEFAULT 10,
    priority INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_connected_at TIMESTAMP WITH TIME ZONE,
    connection_status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    capabilities TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Newsgroups table
CREATE TABLE IF NOT EXISTS newsgroups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    first_article_number BIGINT,
    last_article_number BIGINT,
    last_indexed_article_number BIGINT,
    last_indexed_at TIMESTAMP WITH TIME ZONE,
    last_check_at TIMESTAMP WITH TIME ZONE,
    indexing_priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Articles table
CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,
    message_id VARCHAR(1024) NOT NULL UNIQUE,
    newsgroup_id BIGINT NOT NULL REFERENCES newsgroups(id) ON DELETE CASCADE,
    subject TEXT NOT NULL,
    author VARCHAR(255),
    posted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    article_size BIGINT,
    raw_metadata JSONB,
    normalized_metadata JSONB,
    search_vector tsvector,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create search vector trigger function
CREATE OR REPLACE FUNCTION update_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.subject, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.author, '')), 'D') ||
        setweight(to_tsvector('simple', COALESCE(NEW.normalized_metadata::text, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create search vector trigger
CREATE TRIGGER update_article_search_vector
    BEFORE INSERT OR UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION update_search_vector();

-- Create GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_articles_search_vector ON articles USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_articles_newsgroup_id ON articles (newsgroup_id);
CREATE INDEX IF NOT EXISTS idx_articles_posted_at ON articles (posted_at);
CREATE INDEX IF NOT EXISTS idx_articles_message_id ON articles (message_id);
CREATE INDEX IF NOT EXISTS idx_articles_subject_trgm ON articles USING GIN (subject gin_trgm_ops);

-- Spotnet sources table
CREATE TABLE IF NOT EXISTS spotnet_sources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    source_type VARCHAR(100) NOT NULL,
    configuration JSONB,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Spotnet posts table
CREATE TABLE IF NOT EXISTS spotnet_posts (
    id BIGSERIAL PRIMARY KEY,
    external_id VARCHAR(512) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    author VARCHAR(255),
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    posted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    source_id BIGINT NOT NULL REFERENCES spotnet_sources(id) ON DELETE CASCADE,
    raw_metadata JSONB,
    normalized_metadata JSONB,
    search_vector tsvector,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create search vector trigger for spotnet posts
CREATE TRIGGER update_spotnet_search_vector
    BEFORE INSERT OR UPDATE ON spotnet_posts
    FOR EACH ROW
    EXECUTE FUNCTION update_search_vector();

-- Create GIN index for spotnet search
CREATE INDEX IF NOT EXISTS idx_spotnet_posts_search_vector ON spotnet_posts USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_spotnet_posts_source_id ON spotnet_posts (source_id);
CREATE INDEX IF NOT EXISTS idx_spotnet_posts_posted_at ON spotnet_posts (posted_at);
CREATE INDEX IF NOT EXISTS idx_spotnet_posts_external_id ON spotnet_posts (external_id);
CREATE INDEX IF NOT EXISTS idx_spotnet_posts_title_trgm ON spotnet_posts USING GIN (title gin_trgm_ops);

-- Indexing jobs table
CREATE TABLE IF NOT EXISTS indexing_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_type VARCHAR(100) NOT NULL,
    source_id BIGINT,
    newsgroup_id BIGINT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 0,
    progress DOUBLE PRECISION NOT NULL DEFAULT 0,
    records_processed BIGINT NOT NULL DEFAULT 0,
    records_failed BIGINT NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for active jobs
CREATE INDEX IF NOT EXISTS idx_indexing_jobs_status ON indexing_jobs (status) WHERE status IN ('pending', 'running');

-- Indexing checkpoints table
CREATE TABLE IF NOT EXISTS indexing_checkpoints (
    id BIGSERIAL PRIMARY KEY,
    source_type VARCHAR(100) NOT NULL,
    source_id BIGINT NOT NULL,
    newsgroup_id BIGINT,
    last_processed_position VARCHAR(512) NOT NULL,
    last_processed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_checkpoints_unique UNIQUE(source_type, source_id, newsgroup_id)
);

-- User search history table
CREATE TABLE IF NOT EXISTS user_search_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    search_query TEXT NOT NULL,
    filters JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_search_history_user_id ON user_search_history (user_id);
CREATE INDEX IF NOT EXISTS idx_user_search_history_created_at ON user_search_history (created_at);

-- Bookmarks table
CREATE TABLE IF NOT EXISTS bookmarks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id BIGINT NOT NULL,
    item_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, item_id, item_type)
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks (user_id);

-- System settings table
CREATE TABLE IF NOT EXISTS system_settings (
    id BIGSERIAL PRIMARY KEY,
    setting_key VARCHAR(255) NOT NULL UNIQUE,
    setting_value JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_nntp_servers_updated_at BEFORE UPDATE ON nntp_servers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_newsgroups_updated_at BEFORE UPDATE ON newsgroups FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_spotnet_sources_updated_at BEFORE UPDATE ON spotnet_sources FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_spotnet_posts_updated_at BEFORE UPDATE ON spotnet_posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_indexing_checkpoints_updated_at BEFORE UPDATE ON indexing_checkpoints FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create initial admin user (password: admin)
-- Password hash generated with Argon2id: $argon2id$v=19$m=65536,t=3,p=4$...
INSERT INTO users (username, email, password_hash, role, is_active)
VALUES ('admin', 'admin@localhost', '$argon2id$v=19$m=65536,t=3,p=4$Y2hhbmdlLW1l-in-production$placeholder-hash-for-initial-setup', 'admin', true)
ON CONFLICT (username) DO NOTHING;
