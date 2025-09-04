package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/utils/converter"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
	apperrors "github.com/onefeed-th/onefeed-th-backend-api/internal/errors"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
	"github.com/redis/go-redis/v9"
)

type NewsService interface {
	GetNews(ctx context.Context, req dto.NewsListGetRequest) ([]dto.NewsListGetResponse, error)
	RemoveOldNews(ctx context.Context, req dto.BlankRequest) (any, error)
}

func (s *service) GetNews(ctx context.Context, req dto.NewsListGetRequest) ([]dto.NewsListGetResponse, error) {
	if len(req.Source) == 0 {
		return nil, apperrors.New(apperrors.ValidationError, "source is required").
			WithCode("MISSING_SOURCE").
			WithCaller()
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	var responses []dto.NewsListGetResponse
	redisKey := fmt.Sprintf("news:source=%v:page=%d:limit=%d", req.Source, req.Page, req.Limit)

	slog.Debug("Starting news retrieval",
		"sources", req.Source,
		"page", converter.Int32ToInt(req.Page),
		"limit", converter.Int32ToInt(req.Limit),
		"cache_key", redisKey,
	)

	// Try to get from cache first
	err := s.redis.Get(ctx, redisKey, &responses)
	if err == nil && len(responses) > 0 {
		// Cache hit - return cached data
		slog.Info("Cache hit",
			"cache_key", redisKey,
			"items_count", len(responses),
		)
		return responses, nil
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		// Continue to database query on Redis error, but wrap error for monitoring
		apperrors.Wrap(err, apperrors.RedisError, "failed to retrieve from cache").
			WithCode("CACHE_GET_FAILED").
			WithDetails(fmt.Sprintf("key: %s", redisKey))

		slog.Warn("Cache retrieval failed, continuing with database query",
			"cache_key", redisKey,
			"error_code", "CACHE_GET_FAILED",
			"error", err,
		)
		// Don't return the error (fail gracefully)
	}

	// Cache miss or error - query database
	slog.Debug("Cache miss, querying database",
		"cache_key", redisKey,
	)

	news, err := s.repo.NewsRepository.GetNews(ctx, onefeed_th_sqlc.ListNewsParams{
		Sources:    req.Source,
		PageOffset: (req.Page - 1) * req.Limit,
		PageLimit:  req.Limit,
	})
	if err != nil {
		slog.Error("Database query failed",
			"sources", req.Source,
			"page", req.Page,
			"limit", req.Limit,
			"offset", (req.Page-1)*req.Limit,
			"error", err,
		)
		return nil, apperrors.Wrap(err, apperrors.DatabaseError, "failed to retrieve news from database").
			WithCode("DB_QUERY_FAILED").
			WithDetails(fmt.Sprintf("sources: %v, page: %d, limit: %d", req.Source, req.Page, req.Limit)).
			WithCaller()
	}

	// Build response from database data
	responses = make([]dto.NewsListGetResponse, 0, len(news))
	for _, item := range news {
		responses = append(responses, dto.NewsListGetResponse{
			Title:       item.Title,
			Source:      item.Source,
			PublishedAt: converter.PGTypeTimestampToTime(item.PublishDate),
			Link:        item.Link,
			Image:       item.ImageUrl.String,
		})
	}

	// Cache the result for future requests
	err = s.redis.Set(ctx, redisKey, responses)
	if err != nil {
		slog.Warn("Failed to cache news data",
			"cache_key", redisKey,
			"items_count", len(responses),
			"error_code", "CACHE_SET_FAILED",
			"error", err,
		)
		// Don't fail the request if caching fails
	} else {
		slog.Debug("Successfully cached news data",
			"cache_key", redisKey,
			"items_count", len(responses),
		)
	}

	slog.Info("News retrieval completed",
		"sources", req.Source,
		"page", converter.Int32ToInt(req.Page),
		"limit", converter.Int32ToInt(req.Limit),
		"items_count", len(responses),
	)

	return responses, nil
}

func (s *service) RemoveOldNews(ctx context.Context, req dto.BlankRequest) (any, error) {
	slog.Info("Starting old news removal",
		"retention_days", 30,
	)

	err := s.repo.NewsRepository.RemoveNewsByPublishedDate(ctx)
	if err != nil {
		slog.Error("Failed to remove old news",
			"retention_days", 30,
			"error", err,
		)
		return nil, apperrors.Wrap(err, apperrors.DatabaseError, "failed to remove old news").
			WithCode("DB_DELETE_FAILED").
			WithCaller()
	}

	slog.Info("Successfully removed old news",
		"retention_days", 30,
	)
	return nil, nil
}
