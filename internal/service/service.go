package service

import (
	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/rds"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/repository"
)

type Service interface {
	ServerService
	CollectorService
	NewsService
	TagService
}

type service struct {
	repo  *repository.Repository
	redis rds.RedisClient
}

func NewService(repo *repository.Repository) Service {
	return &service{
		repo:  repo,
		redis: rds.NewRedisClient(),
	}
}
