package service

import (
	"context"
	"fmt"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/utils/converter"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
)

type NewsService interface {
	GetNews(ctx context.Context, req dto.NewsGetRequest) ([]dto.NewsGetResponse, error)
	RemoveOldNews(ctx context.Context, req dto.BlankRequest) (any, error)
}

func (s *service) GetNews(ctx context.Context, req dto.NewsGetRequest) ([]dto.NewsGetResponse, error) {
	if len(req.Source) == 0 {
		return nil, fmt.Errorf("invalid request: source is required")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	news, err := s.repo.NewsRepository.GetNews(ctx, onefeed_th_sqlc.ListNewsParams{
		Sources:    req.Source,
		PageOffset: (req.Page - 1) * req.Limit,
		PageLimit:  req.Limit,
	})
	if err != nil {
		return nil, err
	}

	var responses []dto.NewsGetResponse
	for _, item := range news {
		responses = append(responses, dto.NewsGetResponse{
			Title:       item.Title,
			Source:      item.Source,
			PublishedAt: converter.PGTypeTimestampToTime(item.PublishDate),
			Link:        item.Link,
		})
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
