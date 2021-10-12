package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"hw-async/domain"
	"hw-async/generator"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func initCandle(c *domain.Candle, p domain.Price, cp domain.CandlePeriod) {
	c.Ticker = p.Ticker
	c.Open = p.Value
	c.Close = p.Value
	c.High = p.Value
	c.Low = p.Value
	c.TS = domain.PeriodTS(cp, p.TS)
}

func updateCandle(c *domain.Candle, p domain.Price) {
	c.Close = p.Value
	if p.Value > c.High {
		c.High = p.Value
	}
	if p.Value < c.Low {
		c.Low = p.Value
	}
}

func buildCandle(c []domain.Candle, per domain.CandlePeriod) (*domain.Candle, bool) {
	if len(c) == 0 {
		return nil, false
	}

	max := c[0].High
	min := c[0].Low
	for _, val := range c {
		if val.High > max {
			max = val.High
		}
		if val.Low < min {
			min = val.Low
		}
	}
	return &domain.Candle{
		Ticker: c[0].Ticker,
		Period: per,
		Open:   c[0].Open,
		High:   max,
		Low:    min,
		Close:  c[len(c)-1].Close,
		TS:     domain.PeriodTS(per, c[0].TS),
	}, true
}

func toRecord(c domain.Candle) []string {
	record := make([]string, 0)
	record = append(record, c.Ticker)
	record = append(record, c.TS.String())
	record = append(record, fmt.Sprintf("%f", c.Open))
	record = append(record, fmt.Sprintf("%f", c.High))
	record = append(record, fmt.Sprintf("%f", c.Low))
	record = append(record, fmt.Sprintf("%f", c.Close))

	return record
}

// Цепочка конвеера, записывает свечи в файл и отправляет дальше.
func writeCandle(in <-chan domain.Candle, w *csv.Writer, mu sync.Locker) <-chan domain.Candle {
	out := make(chan domain.Candle)
	go func() {
		for {
			candle, ok := <-in

			// Канал закрылся
			if !ok {
				break
			}

			mu.Lock()
			if err := w.Write(toRecord(candle)); err != nil {
				panic(err)
			}
			mu.Unlock()
			out <- candle
		}
		close(out)
	}()

	return out
}

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

// Начало конвеера, формирует 1-минутные свечи.
func oneMin(prices <-chan domain.Price) <-chan domain.Candle {
	// Канал, по которому далее в конвеер передаются 1-минутные свечи.
	out := make(chan domain.Candle)

	// мапа хранит свечи для каждой комапании.
	mCandles := make(map[string]*domain.Candle)
	for _, ticker := range tickers {
		mCandles[ticker] = &domain.Candle{Period: domain.CandlePeriod1m}
	}

	go func() {
		for price := range prices {
			switch mCandles[price.Ticker].TS {
			// Ввод только начался
			case time.Time{}:
				initCandle(mCandles[price.Ticker], price, domain.CandlePeriod1m)

			// Формируем свечу
			case price.TS:
				updateCandle(mCandles[price.Ticker], price)

			// Свеча закрывается, отправляем по каналу дальше
			default:
				out <- *mCandles[price.Ticker]
				initCandle(mCandles[price.Ticker], price, domain.CandlePeriod1m)
			}
		}
		// Канал закрылся
		// Отправялем все имеющиеся свечи
		for _, val := range mCandles {
			out <- *val
		}
		// Закрываем канал
		close(out)
	}()
	return out
}

// Внутренняя функция конвеера, формирует 2-минутные и 10-минутные свечи.
func intermediateCandle(in <-chan domain.Candle, per domain.CandlePeriod) <-chan domain.Candle {
	out := make(chan domain.Candle)

	// Количество свечей для формирования новой
	var n int
	switch per {
	case domain.CandlePeriod2m:
		n = 2
	case domain.CandlePeriod10m:
		n = 5
	}

	// Хранилище для предыдущих свечей
	mPrevCandles := make(map[string][]domain.Candle)

	// Хранилище для текущих свечей
	mCurrCandle := make(map[string]*domain.Candle)
	for _, ticker := range tickers {
		mCurrCandle[ticker] = &domain.Candle{Period: per}
	}

	go func() {
		for candle := range in {
			mPrevCandles[candle.Ticker] = append(mPrevCandles[candle.Ticker], candle)
			// Если собрали n предыдущих свечей, формируем новую
			if len(mPrevCandles[candle.Ticker]) == n {
				mCurrCandle[candle.Ticker], _ = buildCandle(mPrevCandles[candle.Ticker], per)
				out <- *mCurrCandle[candle.Ticker]
				mPrevCandles[candle.Ticker] = nil
			}
		}
		// Канал закрылся
		// Из имеющихся свечей нужно сформировать новые.
		for _, val := range mCurrCandle {
			cdl, ok := buildCandle(mPrevCandles[val.Ticker], per)
			// Только не пустые свечи.
			if ok {
				out <- *cdl
			}
		}
		close(out)
	}()

	return out
}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  15,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})
	// Канал из которого поступают цены
	price := pg.Prices(ctx)

	// Файл для записи 1-минутных свечей.
	f1, _ := os.Create("candles_1m.csv")
	defer f1.Close()
	w1 := csv.NewWriter(f1)
	defer w1.Flush()
	mu1 := &sync.Mutex{}

	// Файл для записи 2-минутных свечей.
	f2, _ := os.Create("candles_2m.csv")
	defer f2.Close()
	w2 := csv.NewWriter(f2)
	defer w2.Flush()
	mu2 := &sync.Mutex{}

	// Файл для записи 10-минутных свечей.
	f10, _ := os.Create("candles_10m.csv")
	defer f10.Close()
	w10 := csv.NewWriter(f10)
	defer w10.Flush()
	mu10 := &sync.Mutex{}

	// канал для обратоки сигнала SIGINT
	tech := make(chan os.Signal, 1)
	signal.Notify(tech, syscall.SIGINT)

	// Запускаем конвеер
	oneMinCandle := oneMin(price)
	out1 := writeCandle(oneMinCandle, w1, mu1)
	twoMinCandle := intermediateCandle(out1, domain.CandlePeriod2m)
	out2 := writeCandle(twoMinCandle, w2, mu2)
	tenMinCandle := intermediateCandle(out2, domain.CandlePeriod10m)
	outLust := writeCandle(tenMinCandle, w10, mu10)
	for {
		select {
		// Ждем завершения пайплайна
		case <-tech:
			cancelFunc()

			// Нужно осушить пайплайн
			for {
				_, ok := <-outLust
				if !ok {
					return
				}
			}

		case <-outLust:
			// Continue
		}
	}
}
