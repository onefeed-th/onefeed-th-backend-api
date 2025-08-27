package service

import (
	"context"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
)

type TagService interface {
	GetAllTags(ctx context.Context, req dto.BlankRequest) ([]string, error)
}

func (s *service) GetAllTags(ctx context.Context, req dto.BlankRequest) ([]string, error) {
	return s.repo.NewsRepository.GetAllSource(ctx)
}
