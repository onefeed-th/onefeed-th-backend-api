package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
)

type SourceRepository interface {
	GetAllSources(ctx context.Context) ([]onefeed_th_sqlc.Source, error)
}

type SourceRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewSourceRepository(pool *pgxpool.Pool) SourceRepository {
	return &SourceRepositoryImpl{
		pool: pool,
	}
}

func (r *SourceRepositoryImpl) GetAllSources(ctx context.Context) ([]onefeed_th_sqlc.Source, error) {
	query := onefeed_th_sqlc.New(r.pool)
	return query.GetAllSources(ctx)
}
