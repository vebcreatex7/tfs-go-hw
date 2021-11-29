package indicators

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tfs-go-hw/course_project/internal/domain"
)

func TestSMA_Testify(t *testing.T) {
	c := make([]domain.Candle, 3)

	c[0] = domain.Candle{
		Time:  1638007200000,
		Open:  "69000.00000000000",
		High:  "69100.00000000000",
		Low:   "54830.00000000000",
		Close: "55000.00000000000",
	}
	c[1] = domain.Candle{
		Time:  1638010800000,
		Open:  "55000.00000000000",
		High:  "69100.00000000000",
		Low:   "54830.00000000000",
		Close: "54830.00000000000",
	}
	c[2] = domain.Candle{
		Time:  1638014400000,
		Open:  "54830.00000000000",
		High:  "69200.00000000000",
		Low:   "54830.00000000000",
		Close: "69200.00000000000",
	}

	var (
		open  float64
		high  float64
		low   float64
		close float64
	)

	for i := range c {
		price, _ := strconv.ParseFloat(c[i].Open, 64)
		open += price

		price, _ = strconv.ParseFloat(c[i].High, 64)
		high += price

		price, _ = strconv.ParseFloat(c[i].Low, 64)
		low += price

		price, _ = strconv.ParseFloat(c[i].Close, 64)
		close += price
	}
	n := len(c)
	open /= float64(n)
	high /= float64(n)
	low /= float64(n)
	close /= float64(n)
	expect := domain.Candle{
		Open:  fmt.Sprintf("%f", open),
		High:  fmt.Sprintf("%f", high),
		Low:   fmt.Sprintf("%f", low),
		Close: fmt.Sprintf("%f", close),
	}

	assert.Equal(t, expect, SMA(c, n))
}
