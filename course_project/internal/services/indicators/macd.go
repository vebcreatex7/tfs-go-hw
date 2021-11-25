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
	ErrInitMACDNumber = errors.New("MACD Error: Incorrect number of candles")
)

type Macd struct {
	macd         domain.Candle
	macdPrev     domain.Candle
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
	Serve(eg *errgroup.Group, c <-chan domain.Candle) <-chan domain.Action
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
		m.macd = FastSlowDelta(m.fast, m.slow)
		macdPrevs[i] = m.macd
	}
	m.signalPrev = SMA(macdPrevs, m.signalLength)
	m.macdPrev = m.macd

	log.Println("init MACD", m.macdPrev)
	log.Println("init signal", m.signalPrev)
	return nil
}

// Nummber of candles needed to init indicator
func (m *Macd) CandlesNeeded() int {
	return m.slowLength + m.signalLength
}

func (m *Macd) predict() domain.Action {
	var (
		macdPricePrev   float64
		macdPrice       float64
		signalPricePrev float64
		signalPrice     float64
	)
	switch m.source {
	case 'O':
		macdPricePrev, _ = strconv.ParseFloat(m.macdPrev.Open, 64)
		macdPrice, _ = strconv.ParseFloat(m.macd.Open, 64)
		signalPricePrev, _ = strconv.ParseFloat(m.signalPrev.Open, 64)
		signalPrice, _ = strconv.ParseFloat(m.signal.Open, 64)
	case 'H':
		macdPricePrev, _ = strconv.ParseFloat(m.macdPrev.High, 64)
		macdPrice, _ = strconv.ParseFloat(m.macd.High, 64)
		signalPricePrev, _ = strconv.ParseFloat(m.signalPrev.High, 64)
		signalPrice, _ = strconv.ParseFloat(m.signal.High, 64)
	case 'L':
		macdPricePrev, _ = strconv.ParseFloat(m.macdPrev.Low, 64)
		macdPrice, _ = strconv.ParseFloat(m.macd.Low, 64)
		signalPricePrev, _ = strconv.ParseFloat(m.signalPrev.Low, 64)
		signalPrice, _ = strconv.ParseFloat(m.signal.Low, 64)
	case 'C':
		macdPricePrev, _ = strconv.ParseFloat(m.macdPrev.Close, 64)
		macdPrice, _ = strconv.ParseFloat(m.macd.Close, 64)
		signalPricePrev, _ = strconv.ParseFloat(m.signalPrev.Close, 64)
		signalPrice, _ = strconv.ParseFloat(m.signal.Close, 64)
	}

	if macdPricePrev > signalPricePrev && macdPrice < signalPrice {
		return domain.Sell
	} else if macdPricePrev < signalPricePrev && macdPrice > signalPrice {
		return domain.Buy
	} else {
		return domain.Hold
	}
}

func (m *Macd) Serve(eg *errgroup.Group, c <-chan domain.Candle) <-chan domain.Action {
	action := make(chan domain.Action)
	eg.Go(func() error {
		defer func() {
			close(action)
		}()
		for candle := range c {
			fmt.Println(candle)
			m.fast = EMA(candle, m.fastPrev, m.fastLength)
			m.slow = EMA(candle, m.slowPrev, m.slowLength)
			m.macd = FastSlowDelta(m.fast, m.slow)
			m.signal = EMA(m.macd, m.signalPrev, m.signalLength)

			// logic
			fmt.Println("MACD", m.macd)
			fmt.Println("signal", m.signal)

			action <- m.predict()

			m.fastPrev = m.fast
			m.slowPrev = m.slow
			m.macdPrev = m.macd
			m.signalPrev = m.signal

		}
		return nil
	})
	return action
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
