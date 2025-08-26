CREATE TABLE news (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  link TEXT NOT NULL UNIQUE,
  source TEXT NOT NULL,
  image_url TEXT,
  publish_date TIMESTAMP,
  fetched_at TIMESTAMP DEFAULT NOW() -- เวลาเราดึงมาเก็บ
);