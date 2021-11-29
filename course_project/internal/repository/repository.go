package repository

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tfs-go-hw/course_project/internal/repository/queries"
)

type Repo struct {
	*queries.Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repo {
	return &Repo{
		Queries: queries.New(pool),
		pool:    pool,
	}
}
