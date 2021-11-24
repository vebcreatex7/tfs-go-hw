package indicators

import (
	"fmt"

	"github.com/tfs-go-hw/course_project/internal/domain"
	"golang.org/x/sync/errgroup"
)

type Macd struct {
}

type MacdService interface {
	Serve(*errgroup.Group, <-chan domain.Candle)
}

func NewMacd() MacdService {
	return &Macd{}
}

func (m *Macd) Serve(eg *errgroup.Group, c <-chan domain.Candle) {
	eg.Go(func() error {
		for candle := range c {
			fmt.Println(candle)
		}
		return nil
	})
}
