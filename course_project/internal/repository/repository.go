package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository/queries"
)

type repo struct {
	*queries.Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repo{
		Queries: queries.New(pool),
		pool:    pool,
	}
}

type Repository interface {
	InsertOrder(ctx context.Context, order domain.RecordOrder) error //postgres
}
