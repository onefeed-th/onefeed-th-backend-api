package service

import (
	"context"

	"github.com/onefeed-th/onefeed-th-backend-api/internal/core/utils/converter"
	"github.com/onefeed-th/onefeed-th-backend-api/internal/dto"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
)

type SourceService interface {
	GetAllSourceByPagination(ctx context.Context, req dto.GetAllSourceByPaginationRequest) ([]dto.GetAllSourceByPaginationResponse, error)
	CreateSource(ctx context.Context, req dto.CreateSourceRequest) (dto.CreateSourceResponse, error)
}

func (s *service) GetAllSourceByPagination(ctx context.Context, req dto.GetAllSourceByPaginationRequest) ([]dto.GetAllSourceByPaginationResponse, error) {
	sources, err := s.repo.SourceRepository.GetAllSourcesWithPagination(ctx, onefeed_th_sqlc.GetAllSourcesWithPaginationParams{
		PageLimit:  req.PageLimit,
		PageOffset: req.PageOffset,
	})
	if err != nil {
		return nil, err
	}
	var res []dto.GetAllSourceByPaginationResponse
	for _, source := range sources {
		res = append(res, dto.GetAllSourceByPaginationResponse{
			Sources: []dto.Source{
				{
					ID:     int64(source.ID),
					Name:   source.Name,
					Tags:   converter.PGTypeTextToString(source.Tags),
					RSSURL: converter.PGTypeTextToString(source.RssUrl),
				},
			},
		})
	}
	return res, nil
}

func (s *service) CreateSource(ctx context.Context, req dto.CreateSourceRequest) (dto.CreateSourceResponse, error) {
	source, err := s.repo.SourceRepository.CreateSource(ctx, onefeed_th_sqlc.CreateSourceParams{
		Name:   req.Name,
		Tags:   converter.StringToPGTypeTextNull(req.Tags),
		RssUrl: converter.StringToPGTypeTextNull(req.RSSURL),
	})
	if err != nil {
		return dto.CreateSourceResponse{}, err
	}
	return dto.CreateSourceResponse{
		ID:     int64(source.ID),
		Name:   source.Name,
		Tags:   converter.PGTypeTextToString(source.Tags),
		RSSURL: converter.PGTypeTextToString(source.RssUrl),
	}, nil
}
