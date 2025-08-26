package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
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
	sources, err := s.repo.SourceRepository.GetAllSources(ctx)
	if err != nil {
		return dto.Response{}, err
	}

	var newsItems []bulkInsertNewsParams
	var wg sync.WaitGroup
	var mu sync.Mutex

	parser := gofeed.NewParser()

	log.Println("Collecting news from:", len(sources), "sources")
	for _, source := range sources {
		wg.Add(1)
		go func(src onefeed_th_sqlc.Source) {
			defer wg.Done()

			feeds, err := parser.ParseURL(src.RssUrl.String)
			if err != nil {
				log.Println("Error fetching/parsing RSS feed from", src.RssUrl.String, ":", err)
				return
			}

			var localItems []bulkInsertNewsParams
			for _, item := range feeds.Items {
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

	// wait for all go routines
	wg.Wait()

	// insert into database
	err = s.insertNewsWithBatch(ctx, newsItems)
	if err != nil {
		log.Println("Error inserting news items into database:", err)
		return nil, err
	}

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
	for i := 0; i < len(newsItems); i += 100 {
		end := min(i+100, len(newsItems))
		batch := newsItems[i:end]

		// Build query string
		var sb strings.Builder
		args := []interface{}{}
		sb.WriteString(`INSERT INTO news (title, link, source, image_url, publish_date, fetched_at) VALUES `)

		for j, item := range batch {
			argPos := j*5 + 1
			sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,NOW())",
				argPos, argPos+1, argPos+2, argPos+3, argPos+4))
			if j < len(batch)-1 {
				sb.WriteString(",")
			}

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
