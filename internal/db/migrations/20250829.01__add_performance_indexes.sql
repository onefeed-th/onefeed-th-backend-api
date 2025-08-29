-- Add performance indexes for news table

-- Index for source filtering (used in ListNews query)
CREATE INDEX IF NOT EXISTS idx_news_source ON news(source);

-- Composite index for source + publish_date (optimal for ListNews query with ORDER BY)
CREATE INDEX IF NOT EXISTS idx_news_source_publish_date ON news(source, publish_date DESC);

-- Index for publish_date (used in RemoveNewsByPublishedDate and ordering)
CREATE INDEX IF NOT EXISTS idx_news_publish_date ON news(publish_date DESC);

-- Index for fetched_at (useful for maintenance queries)
CREATE INDEX IF NOT EXISTS idx_news_fetched_at ON news(fetched_at DESC);

-- Partial index for recent news (optimize common queries for recent content)
CREATE INDEX IF NOT EXISTS idx_news_recent ON news(source, publish_date DESC)
WHERE publish_date >= NOW() - INTERVAL '7 days';