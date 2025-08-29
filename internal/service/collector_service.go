package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/logger"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
	"go.uber.org/zap"
)

type CollectorService interface {
	CollectNewsFromSource(ctx context.Context, req dto.BlankRequest) (any, error)
}

type bulkInsertNewsParams struct {
	Title       string
	Link        string
	Source      string
	ImageUrl    string
	PublishDate *time.Time
}

func (s *service) CollectNewsFromSource(ctx context.Context, req dto.BlankRequest) (any, error) {
	log := logger.New("collector-service")
	
	sources, err := s.repo.SourceRepository.GetAllSources(ctx)
	if err != nil {
		log.Error(ctx, "Failed to get sources", zap.Error(err))
		return dto.Response{}, err
	}

	// Pre-allocate slice with estimated capacity (avg 20 items per source)
	newsItems := make([]bulkInsertNewsParams, 0, len(sources)*20)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	parser := gofeed.NewParser()
	parser.Client = httpClient

	log.Info(ctx, "Starting news collection", 
		zap.Int("source_count", len(sources)),
	)
	
	// Create a context with timeout for the entire collection process
	collectCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for _, source := range sources {
		wg.Add(1)
		go func(src onefeed_th_sqlc.Source) {
			defer wg.Done()

			// Check if context is already cancelled
			select {
			case <-collectCtx.Done():
				log.Warn(ctx, "Context cancelled for source", 
					zap.String("source", src.Name),
					zap.Error(collectCtx.Err()),
				)
				return
			default:
			}

			// Create individual timeout for each RSS feed
			feedCtx, feedCancel := context.WithTimeout(collectCtx, 30*time.Second)
			defer feedCancel()

			feeds, err := parser.ParseURLWithContext(src.RssUrl.String, feedCtx)
			if err != nil {
				log.Error(ctx, "Error parsing RSS feed", 
					zap.String("source", src.Name),
					zap.String("rss_url", src.RssUrl.String),
					zap.Error(err),
				)
				return
			}

			// Pre-allocate local items slice based on feed size
			localItems := make([]bulkInsertNewsParams, 0, len(feeds.Items))
			for _, item := range feeds.Items {
				// Check for cancellation during processing
				select {
				case <-feedCtx.Done():
					log.Warn(ctx, "Feed processing cancelled", 
						zap.String("source", src.Name),
					)
					return
				default:
				}

				news := bulkInsertNewsParams{
					Title:       item.Title,
					Link:        sanitizeLink(item.Link),
					Source:      src.Name,
					ImageUrl:    extractImage(item),
					PublishDate: item.PublishedParsed,
				}
				localItems = append(localItems, news)
			}

			mu.Lock()
			newsItems = append(newsItems, localItems...)
			mu.Unlock()
		}(source)
	}

	// Wait for all goroutines with context timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed normally
		log.Debug(ctx, "All RSS feeds processed successfully")
	case <-collectCtx.Done():
		log.Error(ctx, "Collection timed out", zap.Error(collectCtx.Err()))
		return nil, fmt.Errorf("news collection timed out: %w", collectCtx.Err())
	}

	// insert into database
	log.Info(ctx, "Inserting news items into database", 
		zap.Int("total_items", len(newsItems)),
	)
	
	err = s.insertNewsWithBatch(ctx, newsItems)
	if err != nil {
		log.Error(ctx, "Error inserting news items into database", zap.Error(err))
		return nil, err
	}

	// Clear news cache
	err = s.redis.RemoveKeyContaining(ctx, "news")
	if err != nil {
		log.Error(ctx, "Error removing news cache keys", zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "News collection completed successfully", 
		zap.Int("total_items", len(newsItems)),
		zap.Int("source_count", len(sources)),
	)

	return nil, nil
}

func extractImage(item *gofeed.Item) string {
	if item.Image != nil {
		return item.Image.URL
	}

	if len(item.Enclosures) > 0 {
		return item.Enclosures[0].URL
	}

	html := item.Description
	if html == "" {
		html = item.Content
	}

	if html != "" {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err == nil {
			if imgSrc, exists := doc.Find("img").First().Attr("src"); exists {
				return imgSrc
			}
		}
	}

	return ""
}

func sanitizeLink(raw string) string {
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, "|")
	if len(parts) > 1 {
		return parts[len(parts)-1] // เอาตัวท้ายสุด
	}
	return raw
}

func (s *service) insertNewsWithBatch(ctx context.Context, newsItems []bulkInsertNewsParams) error {
	const batchSize = 100
	
	for i := 0; i < len(newsItems); i += batchSize {
		end := min(i+batchSize, len(newsItems))
		batch := newsItems[i:end]

		// Pre-allocate slice capacity for better memory efficiency
		args := make([]interface{}, 0, len(batch)*5)
		
		// Pre-allocate strings.Builder with estimated capacity
		var sb strings.Builder
		// Estimate: base query + (placeholder chars * items) + commas
		estimatedSize := 80 + (len(batch) * 25) + len(batch)
		sb.Grow(estimatedSize)
		
		sb.WriteString(`INSERT INTO news (title, link, source, image_url, publish_date, fetched_at) VALUES `)

		for j, item := range batch {
			argPos := j*5 + 1
			sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,NOW())",
				argPos, argPos+1, argPos+2, argPos+3, argPos+4))
			if j < len(batch)-1 {
				sb.WriteString(",")
			}

			// Append to pre-allocated slice
			args = append(args,
				item.Title,
				item.Link,
				item.Source,
				item.ImageUrl,
				item.PublishDate,
			)
		}

		sb.WriteString(" ON CONFLICT (link) DO NOTHING;")

		// Exec batch insert
		err := s.repo.NewsRepository.BulkInsertNews(ctx, sb.String(), args)
		if err != nil {
			return fmt.Errorf("batch insert failed: %w", err)
		}
	}

	return nil
}
