package service

import (
	"context"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
)

type ServerService interface {
	HealthCheck(ctx context.Context, req dto.BlankRequest) (string, error)
}

func (s *service) HealthCheck(ctx context.Context, req dto.BlankRequest) (string, error) {
	return "OK", nil
}
