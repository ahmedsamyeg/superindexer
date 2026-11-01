-- +migrate Down
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_nntp_servers_updated_at ON nntp_servers;
DROP TRIGGER IF EXISTS update_newsgroups_updated_at ON newsgroups;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_spotnet_sources_updated_at ON spotnet_sources;
DROP TRIGGER IF EXISTS update_spotnet_posts_updated_at ON spotnet_posts;
DROP TRIGGER IF EXISTS update_indexing_checkpoints_updated_at ON indexing_checkpoints;
DROP TRIGGER IF EXISTS update_article_search_vector ON articles;
DROP TRIGGER IF EXISTS update_spotnet_search_vector ON spotnet_posts;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS update_search_vector();

DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS user_search_history;
DROP TABLE IF EXISTS indexing_checkpoints;
DROP TABLE IF EXISTS indexing_jobs;
DROP TABLE IF EXISTS spotnet_posts;
DROP TABLE IF EXISTS spotnet_sources;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS newsgroups;
DROP TABLE IF EXISTS nntp_servers;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS unaccent;
