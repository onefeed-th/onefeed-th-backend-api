package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	onefeed_th_sqlc "github.com/onefeed-th/onefeed-th-backend-api/internal/sqlc/onefeed_th_sqlc/db"
)

type NewsRepository interface {
	BulkInsertNews(ctx context.Context, stringBuilder string, args []interface{}) error
	GetNews(ctx context.Context, params onefeed_th_sqlc.ListNewsParams) ([]onefeed_th_sqlc.News, error)
	RemoveNewsByPublishedDate(ctx context.Context) error
	GetAllSource(ctx context.Context) ([]string, error)
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

func (r *NewsRepositoryImpl) GetNews(ctx context.Context, params onefeed_th_sqlc.ListNewsParams) ([]onefeed_th_sqlc.News, error) {
	query := onefeed_th_sqlc.New(r.pool)
	return query.ListNews(ctx, params)
}

func (r *NewsRepositoryImpl) RemoveNewsByPublishedDate(ctx context.Context) error {
	query := onefeed_th_sqlc.New(r.pool)
	return query.RemoveNewsByPublishedDate(ctx)
}

func (r *NewsRepositoryImpl) GetAllSource(ctx context.Context) ([]string, error) {
	query := onefeed_th_sqlc.New(r.pool)
	return query.GetAllSource(ctx)
}
