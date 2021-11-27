package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository/queries"
	"github.com/tfs-go-hw/course_project/internal/repository/telegram"
)

type repo struct {
	*queries.Queries
	pool  *pgxpool.Pool
	tgbot *telegram.TgBot
}

func NewRepository(pool *pgxpool.Pool, tgbot *telegram.TgBot) Repository {
	return &repo{
		Queries: queries.New(pool),
		pool:    pool,
		tgbot:   tgbot,
	}
}

func (r *repo) SendOrder(order domain.RecordOrder) error {
	return r.tgbot.SendOrder(order)
}

type Repository interface {
	InsertOrder(ctx context.Context, order domain.RecordOrder) error //postgres
	SendOrder(order domain.RecordOrder) error                        // telegram
}
