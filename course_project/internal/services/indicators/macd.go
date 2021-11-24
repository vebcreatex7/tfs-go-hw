package indicators

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/tfs-go-hw/course_project/internal/domain"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInitMACDNumber = errors.New("Incorrect number of candles")
)

type Macd struct {
	//macd         domain.Candle
	//macdPrev     domain.Candle
	fastPrev     domain.Candle
	fast         domain.Candle
	slowPrev     domain.Candle
	slow         domain.Candle
	signal       domain.Candle
	signalPrev   domain.Candle
	fastLength   int  // shorter period
	slowLength   int  // longer period
	signalLength int  // signal period
	source       rune // O/H/L/C price

}

type MacdService interface {
	Serve(eg *errgroup.Group, c <-chan domain.Candle)
	InitMacd(candles []domain.Candle) error
	CandlesNeeded() int
}

func NewMacd() MacdService {
	return &Macd{
		fastLength:   12,
		slowLength:   26,
		signalLength: 9,
		source:       'C',
	}
}

func (m *Macd) InitMacd(candles []domain.Candle) error {
	if len(candles) != m.slowLength+m.signalLength {
		return ErrInitMACDNumber
	}
	m.fastPrev = SMA(candles[m.slowLength-m.fastLength:m.slowLength], m.fastLength)
	m.slowPrev = SMA(candles[:m.slowLength], m.slowLength)

	macdPrevs := make([]domain.Candle, m.signalLength)
	for i := range macdPrevs {
		m.fast = EMA(candles[i+m.slowLength], m.fastPrev, m.fastLength)
		m.slow = EMA(candles[i+m.slowLength], m.slowPrev, m.slowLength)
		m.fastPrev = m.fast
		m.slowPrev = m.slow
		macd := FastSlowDelta(m.fast, m.slow)
		macdPrevs[i] = macd
	}
	m.signalPrev = SMA(macdPrevs, m.signalLength)

	log.Println("init MACD", FastSlowDelta(m.fastPrev, m.slowPrev))
	log.Println("init signal", m.signalPrev)
	return nil
}

// Nummber of candles needed to init indicator
func (m *Macd) CandlesNeeded() int {
	return m.slowLength + m.signalLength
}

func (m *Macd) Serve(eg *errgroup.Group, c <-chan domain.Candle) {
	eg.Go(func() error {
		for candle := range c {
			fmt.Println(candle)
			m.fast = EMA(candle, m.fastPrev, m.fastLength)
			m.slow = EMA(candle, m.slowPrev, m.slowLength)
			macd := FastSlowDelta(m.fast, m.slow)
			m.signal = EMA(macd, m.signalPrev, m.signalLength)

			// logic
			fmt.Println("MACD", macd)
			fmt.Println("signal", m.signal)

			m.fastPrev = m.fast
			m.slowPrev = m.slow
			m.signalPrev = m.signal

		}
		return nil
	})
}

func FastSlowDelta(fast domain.Candle, slow domain.Candle) domain.Candle {
	var (
		deltaOpen  float64
		deltaHigh  float64
		deltaLow   float64
		deltaClose float64
	)

	fastOpen, _ := strconv.ParseFloat(fast.Open, 64)
	slowOpen, _ := strconv.ParseFloat(slow.Open, 64)
	deltaOpen = fastOpen - slowOpen

	fastHigh, _ := strconv.ParseFloat(fast.High, 64)
	slowHigh, _ := strconv.ParseFloat(slow.High, 64)
	deltaHigh = fastHigh - slowHigh

	fastLow, _ := strconv.ParseFloat(fast.Low, 64)
	slowLow, _ := strconv.ParseFloat(slow.Low, 64)
	deltaLow = fastLow - slowLow

	fastClose, _ := strconv.ParseFloat(fast.Close, 64)
	slowClose, _ := strconv.ParseFloat(slow.Close, 64)
	deltaClose = fastClose - slowClose

	return domain.Candle{
		Open:  fmt.Sprintf("%f", deltaOpen),
		High:  fmt.Sprintf("%f", deltaHigh),
		Low:   fmt.Sprintf("%f", deltaLow),
		Close: fmt.Sprintf("%f", deltaClose),
	}

}
