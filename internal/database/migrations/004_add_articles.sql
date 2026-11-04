-- +goose Up
CREATE TABLE articles (
    id BIGSERIAL PRIMARY KEY,
    server_id BIGINT NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    group_name TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE IF EXISTS articles;
