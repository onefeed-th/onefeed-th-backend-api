package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/utils/converter"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
	"github.com/redis/go-redis/v9"
)

type NewsService interface {
	GetNews(ctx context.Context, req dto.NewsListGetRequest) ([]dto.NewsListGetResponse, error)
	RemoveOldNews(ctx context.Context, req dto.BlankRequest) (any, error)
}

func (s *service) GetNews(ctx context.Context, req dto.NewsListGetRequest) ([]dto.NewsListGetResponse, error) {
	if len(req.Source) == 0 {
		return nil, fmt.Errorf("invalid request: source is required")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	var responses []dto.NewsListGetResponse
	redisKey := fmt.Sprintf("news:source=%v:page=%d:limit=%d", req.Source, req.Page, req.Limit)
	log.Println("Redis Key:", redisKey)

	err := s.redis.Get(ctx, redisKey, &responses)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Println("Redis Get Error:", err)
		return nil, err
	}

	news, err := s.repo.NewsRepository.GetNews(ctx, onefeed_th_sqlc.ListNewsParams{
		Sources:    req.Source,
		PageOffset: (req.Page - 1) * req.Limit,
		PageLimit:  req.Limit,
	})
	if err != nil {
		return nil, err
	}

	for _, item := range news {
		responses = append(responses, dto.NewsListGetResponse{
			Title:       item.Title,
			Source:      item.Source,
			PublishedAt: converter.PGTypeTimestampToTime(item.PublishDate),
			Link:        item.Link,
			Image:       item.ImageUrl.String,
		})
	}

	err = s.redis.Set(ctx, redisKey, responses)
	if err != nil {
		log.Println("Redis Set Error:", err)
	}

	return responses, nil
}

func (s *service) RemoveOldNews(ctx context.Context, req dto.BlankRequest) (any, error) {
	err := s.repo.NewsRepository.RemoveNewsByPublishedDate(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
