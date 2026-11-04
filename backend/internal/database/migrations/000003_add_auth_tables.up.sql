-- Create token blacklist table for JWT token revocation
CREATE TABLE IF NOT EXISTS token_blacklist (
    id          BIGSERIAL PRIMARY KEY,
    token_hash  TEXT NOT NULL,
    user_id     BIGINT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fast token lookup
CREATE INDEX idx_token_blacklist_token_hash ON token_blacklist (token_hash);

-- Index for cleanup queries
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist (expires_at);

-- Create refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    token       TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    is_revoked  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fast token lookup
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens (token);

-- Index for user token lookup
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- Index for cleanup queries
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);
