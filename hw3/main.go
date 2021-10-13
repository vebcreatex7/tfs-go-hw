package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"hw-async/domain"
	"hw-async/generator"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
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
func writeCandle(in <-chan domain.Candle, per domain.CandlePeriod, eg *errgroup.Group) <-chan domain.Candle {
	out := make(chan domain.Candle)
	eg.Go(func() error {
		f, _ := os.Create(fmt.Sprintf("candles_%s.csv", per))
		w := csv.NewWriter(f)
		defer f.Close()
		defer w.Flush()
		defer close(out)

		for candle := range in {
			if err := w.Write(toRecord(candle)); err != nil {
				panic(err)
			}
			out <- candle
		}
		return errors.New("channel is closed")
	})

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
		defer close(out)
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
	}()
	return out
}

// Внутренняя функция конвеера, формирует 2-минутные и 10-минутные свечи.
func intermediateCandle(in <-chan domain.Candle, per domain.CandlePeriod) <-chan domain.Candle {
	out := make(chan domain.Candle)

	// Хранилище для предыдущих свечей
	mPrevCandles := make(map[string][]domain.Candle)

	// Хранилище для текущих свечей
	mCurrCandle := make(map[string]*domain.Candle)
	for _, ticker := range tickers {
		mCurrCandle[ticker] = &domain.Candle{Period: per}
	}

	go func() {
		defer close(out)
		for candle := range in {
			n := len(mPrevCandles[candle.Ticker])

			// Если у нас 0 накопленных свечей или период новой свечи совпадает с периодом накопленных
			if n == 0 || (domain.PeriodTS(per, mPrevCandles[candle.Ticker][n-1].TS) == domain.PeriodTS(per, candle.TS)) {
				mPrevCandles[candle.Ticker] = append(mPrevCandles[candle.Ticker], candle)

				// Формируем новую свечку
			} else {
				mCurrCandle[candle.Ticker], _ = buildCandle(mPrevCandles[candle.Ticker], per)
				out <- *mCurrCandle[candle.Ticker]
				mPrevCandles[candle.Ticker] = nil
			}
		}
		// Канал закрылся
		// Из имеющихся свечей нужно сформировать новые.
		for s := range mCurrCandle {
			cdl, ok := buildCandle(mPrevCandles[s], per)
			// Только не пустые свечи.
			if ok {
				out <- *cdl
			}
		}
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

	// канал для обратоки сигнала SIGINT
	tech := make(chan os.Signal, 1)
	signal.Notify(tech, syscall.SIGINT)

	var eg errgroup.Group

	// Запускаем конвеер

	oneMinCandle := oneMin(price)
	out1 := writeCandle(oneMinCandle, domain.CandlePeriod1m, &eg)
	twoMinCandle := intermediateCandle(out1, domain.CandlePeriod2m)
	out2 := writeCandle(twoMinCandle, domain.CandlePeriod2m, &eg)
	tenMinCandle := intermediateCandle(out2, domain.CandlePeriod10m)
	outLust := writeCandle(tenMinCandle, domain.CandlePeriod10m, &eg)
	go func() {
		for range outLust {
		}
	}()

	<-tech
	cancelFunc()
	_ = eg.Wait()

}
