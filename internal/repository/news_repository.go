package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NewsRepository interface {
	BulkInsertNews(ctx context.Context, stringBuilder string, args []interface{}) error
}

type NewsRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewNewsRepository(pool *pgxpool.Pool) NewsRepository {
	return &NewsRepositoryImpl{
		pool: pool,
	}
}

func (r *NewsRepositoryImpl) BulkInsertNews(ctx context.Context, stringBuilder string, args []interface{}) error {
	_, err := r.pool.Exec(ctx, stringBuilder, args...)
	if err != nil {
		return err
	}
	return nil
}
