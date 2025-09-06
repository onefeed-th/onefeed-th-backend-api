CREATE TABLE sources (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  tags TEXT NULL,
  rss_url TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);
-- name: GetAllSources :many
SELECT *
FROM sources;
-- name: GetAllSourcesWithPagination :many
SELECT *
FROM sources
ORDER BY created_at DESC
LIMIT @page_limit OFFSET @page_offset;
-- name: CreateSource :one
INSERT INTO sources (name, tags, rss_url)
VALUES (@name, @tags, @rss_url)
RETURNING *;