package indicators

import (
	"fmt"
	"strconv"

	"github.com/tfs-go-hw/course_project/internal/domain"
)

func SMA(candles []domain.Candle, n int) domain.Candle {
	var (
		open  float64
		high  float64
		low   float64
		close float64
	)

	for i := range candles {
		price, _ := strconv.ParseFloat(candles[i].Open, 64)
		open += price

		price, _ = strconv.ParseFloat(candles[i].High, 64)
		high += price

		price, _ = strconv.ParseFloat(candles[i].Low, 64)
		low += price

		price, _ = strconv.ParseFloat(candles[i].Close, 64)
		close += price
	}

	open /= float64(n)
	high /= float64(n)
	low /= float64(n)
	close /= float64(n)

	return domain.Candle{
		Open:  fmt.Sprintf("%f", open),
		High:  fmt.Sprintf("%f", high),
		Low:   fmt.Sprintf("%f", low),
		Close: fmt.Sprintf("%f", close),
	}

}
