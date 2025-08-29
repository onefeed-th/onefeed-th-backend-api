package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/utils/converter"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
	apperrors "github.com/onefeed-th/onefeed-th-backend-api/internal/errors"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/logger"
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

	log := logger.New("news-service")
	log.Debug(ctx, "Starting news retrieval", map[string]interface{}{
		"sources":   req.Source,
		"page":      req.Page,
		"limit":     req.Limit,
		"cache_key": redisKey,
	})

	// Try to get from cache first
	err := s.redis.Get(ctx, redisKey, &responses)
	if err == nil && len(responses) > 0 {
		// Cache hit - return cached data
		log.Info(ctx, "Cache hit", map[string]interface{}{
			"cache_key":   redisKey,
			"items_count": len(responses),
		})
		return responses, nil
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		// Continue to database query on Redis error, but wrap error for monitoring
		apperrors.Wrap(err, apperrors.RedisError, "failed to retrieve from cache").
			WithCode("CACHE_GET_FAILED").
			WithDetails(fmt.Sprintf("key: %s", redisKey))

		log.Warn(ctx, "Cache retrieval failed, continuing with database query", map[string]interface{}{
			"cache_key":  redisKey,
			"error_code": "CACHE_GET_FAILED",
		})
		// Don't return the error (fail gracefully)
	}

	// Cache miss or error - query database
	log.Debug(ctx, "Cache miss, querying database", map[string]interface{}{
		"cache_key": redisKey,
	})

	news, err := s.repo.NewsRepository.GetNews(ctx, onefeed_th_sqlc.ListNewsParams{
		Sources:    req.Source,
		PageOffset: (req.Page - 1) * req.Limit,
		PageLimit:  req.Limit,
	})
	if err != nil {
		log.Error(ctx, "Database query failed", err, map[string]interface{}{
			"sources": req.Source,
			"page":    req.Page,
			"limit":   req.Limit,
			"offset":  (req.Page - 1) * req.Limit,
		})
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
		log.Warn(ctx, "Failed to cache news data", map[string]interface{}{
			"cache_key":   redisKey,
			"items_count": len(responses),
			"error_code":  "CACHE_SET_FAILED",
		})
		// Don't fail the request if caching fails
	} else {
		log.Debug(ctx, "Successfully cached news data", map[string]interface{}{
			"cache_key":   redisKey,
			"items_count": len(responses),
		})
	}

	log.Info(ctx, "News retrieval completed", map[string]interface{}{
		"sources":     req.Source,
		"page":        req.Page,
		"limit":       req.Limit,
		"items_count": len(responses),
	})

	return responses, nil
}

func (s *service) RemoveOldNews(ctx context.Context, req dto.BlankRequest) (any, error) {
	log := logger.New("news-service")
	log.Info(ctx, "Starting old news removal", map[string]interface{}{
		"retention_days": 30,
	})

	err := s.repo.NewsRepository.RemoveNewsByPublishedDate(ctx)
	if err != nil {
		log.Error(ctx, "Failed to remove old news", err, map[string]interface{}{
			"retention_days": 30,
		})
		return nil, apperrors.Wrap(err, apperrors.DatabaseError, "failed to remove old news").
			WithCode("DB_DELETE_FAILED").
			WithCaller()
	}

	log.Info(ctx, "Successfully removed old news", map[string]interface{}{
		"retention_days": 30,
	})
	return nil, nil
}
