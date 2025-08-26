package repository

import "github.com/onefeed-th/onefeed-th-backend-api/internal/db"

type Repository struct {
	SourceRepository SourceRepository
	NewsRepository   NewsRepository
}

func NewRepository() *Repository {
	pool := db.GetPool()

	return &Repository{
		SourceRepository: NewSourceRepository(pool),
		NewsRepository:   NewNewsRepository(pool),
	}
}
