package indicators

import (
	"fmt"
	"strconv"

	"github.com/tfs-go-hw/course_project/internal/domain"
)

/*
func EMA(candle domain.Candle, prevEMA float64, n int, source rune) float64 {

	var price float64 = 0.
	var err error
	switch source {
	case 'O':
		price, err = strconv.ParseFloat(candle.Open, 64)
		if err != nil {
			return 0.
		}
	case 'H':
		price, err = strconv.ParseFloat(candle.High, 64)
		if err != nil {
			return 0.
		}
	case 'L':
		price, err = strconv.ParseFloat(candle.Low, 64)
		if err != nil {
			return 0.
		}
	case 'C':
		price, err = strconv.ParseFloat(candle.Close, 64)
		if err != nil {
			return 0.
		}
	}
	return float64(2)/float64(n+1)*(price-prevEMA) + prevEMA
}
*/

func EMA(candle domain.Candle, prevEMA domain.Candle, n int) domain.Candle {
	var (
		open  float64
		high  float64
		low   float64
		close float64
	)
	price, _ := strconv.ParseFloat(candle.Open, 64)
	prevOpen, _ := strconv.ParseFloat(prevEMA.Open, 64)
	open = float64(2)/float64(n+1)*(price-prevOpen) + prevOpen

	price, _ = strconv.ParseFloat(candle.High, 64)
	prevHigh, _ := strconv.ParseFloat(prevEMA.High, 64)
	high = float64(2)/float64(n+1)*(price-prevHigh) + prevHigh

	price, _ = strconv.ParseFloat(candle.Low, 64)
	prevLow, _ := strconv.ParseFloat(prevEMA.Low, 64)
	low = float64(2)/float64(n+1)*(price-prevLow) + prevLow

	price, _ = strconv.ParseFloat(candle.Close, 64)
	prevClose, _ := strconv.ParseFloat(prevEMA.Close, 64)
	close = float64(2)/float64(n+1)*(price-prevClose) + prevClose

	return domain.Candle{
		Open:  fmt.Sprintf("%f", open),
		High:  fmt.Sprintf("%f", high),
		Low:   fmt.Sprintf("%f", low),
		Close: fmt.Sprintf("%f", close),
	}
}
