package queries

import (
	"context"

	"github.com/tfs-go-hw/course_project/internal/domain"
)

const addOrder = `INSERT INTO journal (date, symbol, side, size, price)
VALUES ($1, $2, $3, $4, $5);`

func (q *Queries) InsertOrder(ctx context.Context, order domain.RecordOrder) error {
	_, err := q.pool.Exec(ctx, addOrder, order.TS, order.Symbol, order.Side, order.Size, order.Price)
	return err
}
