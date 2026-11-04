-- +goose Up
-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    password_reset_token TEXT,
    password_reset_expiry TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create nntp_servers table
CREATE TABLE nntp_servers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    hostname TEXT NOT NULL,
    port INT NOT NULL,
    username TEXT,
    encrypted_password TEXT,
    tls_enabled BOOLEAN NOT NULL DEFAULT false,
    max_connections INT NOT NULL DEFAULT 10,
    priority INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_connected_at TIMESTAMPTZ,
    connection_status TEXT NOT NULL DEFAULT 'unknown',
    capabilities TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create newsgroups table
CREATE TABLE newsgroups (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    first_article_number BIGINT,
    last_article_number BIGINT,
    last_indexed_article_number BIGINT,
    last_indexed_at TIMESTAMPTZ,
    last_check_at TIMESTAMPTZ,
    indexing_priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create articles table
CREATE TABLE articles (
    id BIGSERIAL PRIMARY KEY,
    message_id TEXT NOT NULL UNIQUE,
    newsgroup_id BIGINT NOT NULL REFERENCES newsgroups(id) ON DELETE CASCADE,
    subject TEXT NOT NULL,
    author TEXT,
    posted_at TIMESTAMPTZ NOT NULL,
    article_size BIGINT,
    raw_metadata JSONB,
    normalized_metadata JSONB,
    search_vector TSVECTOR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS newsgroups;
DROP TABLE IF EXISTS nntp_servers;
DROP TABLE IF EXISTS users;
