DROP TABLE IF EXISTS sources;
CREATE TABLE sources (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  tags TEXT NULL,
  rss_url TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);