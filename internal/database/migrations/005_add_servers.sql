-- +goose Up
CREATE TABLE servers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    host TEXT NOT NULL,
    port INT NOT NULL,
    username TEXT,
    password TEXT
);
-- +goose Down
DROP TABLE IF EXISTS servers;
