CREATE TABLE news (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  link TEXT NOT NULL UNIQUE,
  source TEXT NOT NULL,
  image_url TEXT,
  publish_date TIMESTAMP,
  fetched_at TIMESTAMP DEFAULT NOW() -- เวลาเราดึงมาเก็บ
);
-- name: ListNews :many
SELECT *
FROM news
WHERE news.source = ANY(@sources::TEXT [])
ORDER BY publish_date DESC
LIMIT @page_limit OFFSET @page_offset;
-- name: RemoveNewsByPublishedDate :exec
DELETE FROM news
WHERE publish_date < NOW() - INTERVAL '30 days';
-- name: GetAllSource :many
SELECT DISTINCT source
FROM news;
-- name: GetAllMissingLinks :many
WITH recv AS (
  SELECT unnest(@links::TEXT []) AS link
)
SELECT r.link::TEXT AS missing_link
FROM recv r
  LEFT JOIN news n ON r.link = n.link
WHERE n.link IS NULL;