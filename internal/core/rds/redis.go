package rds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/onefeed-th/onefeed-th-backend-api/config"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func InitRedis(ctx context.Context) error {
	config := config.GetConfig()
	password := config.Redis.Password
	host := config.Redis.Host
	port := config.Redis.Port

	var missing []string
	if host == "" {
		missing = append(missing, "REDIS_HOST")
	}
	if port == 0 {
		missing = append(missing, "REDIS_PORT")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       0, // use default DB
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	return nil
}

func GetClient() *redis.Client {
	return client
}

type RedisClient interface {
	SetWithExpiredTime(ctx context.Context, key string, value any, expiration time.Duration) error
	Set(ctx context.Context, key string, value any) error
	Get(ctx context.Context, key string, dest any) error
	RemoveKeyContaining(ctx context.Context, containKey string) error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient() RedisClient {
	return &redisClient{
		client: GetClient(),
	}
}

func (r *redisClient) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return err
	}
	return nil
}

func (r *redisClient) SetWithExpiredTime(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %q: %w", key, err)
	}
	return nil
}

func (r *redisClient) Set(ctx context.Context, key string, value any) error {
	bytes, err := json.Marshal(value)
	if err == nil {
		if err := r.client.Set(ctx, key, bytes, 0).Err(); err != nil {
			return err
		}
	}
	return err
}

func (r *redisClient) RemoveKeyContaining(ctx context.Context, containKey string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, fmt.Sprintf("*%s*", containKey), 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
