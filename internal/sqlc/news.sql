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
WHERE news.source = ANY(@sources::TEXT[])
ORDER BY publish_date DESC
LIMIT @page_limit OFFSET @page_offset;