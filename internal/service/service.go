package service

import "github.com/onefeed-th/onefeed-th-backend-api/internal/repository"

type Service interface {
	ServerService
	CollectorService
}

type service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) Service {
	return &service{
		repo: repo,
	}
}
