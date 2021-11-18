package queries

import (
	"context"

	"github.com/tfs-go-hw/course_project/internal/domain"
)

const addOrder = ""

func (q *Queries) InsertOrder(ctx context.Context, order domain.Order) error {
	_, err := q.pool.Exec(ctx, addOrder)
	return err
}
