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

func buildCandle(c []domain.Candle, per domain.CandlePeriod) *domain.Candle {
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
	}
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

func initMapOfCandle(mCandles map[string]*domain.Candle, per domain.CandlePeriod) {
	for _, ticker := range tickers {
		mCandles[ticker] = &domain.Candle{Period: per}
	}
}

// Цепочка конвеера, записывает свечи в файл и отправляет дальше.
func writeCandle(in <-chan domain.Candle, w *csv.Writer) <-chan domain.Candle {
	out := make(chan domain.Candle)
	x := <-in
	if err := w.Write(toRecord(x)); err != nil {
		panic(err)
	}
	out <- x
	return out
}

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

func oneMin(ctx context.Context, pg *generator.PricesGenerator, wg *sync.WaitGroup, cancel context.CancelFunc, sync chan struct{}) <-chan domain.Candle {
	// По каналу идут сгенерированные цены.
	prices := pg.Prices(ctx)

	// Канал, по которому передаются 1-минутные свечи.
	out := make(chan domain.Candle)

	// мапа хранит свечи для каждой комапании.

	mCandles := make(map[string]*domain.Candle)
	initMapOfCandle(mCandles, domain.CandlePeriod1m)

	// канал для обратоки сигнала SIGINT
	tech := make(chan os.Signal)
	signal.Notify(tech, syscall.SIGINT)

	nullTime := time.Time{}
	go func() {
		f, _ := os.Create("candles_1m.csv")
		defer f.Close()

		w := csv.NewWriter(f)
		defer w.Flush()

		defer wg.Done()
		defer close(tech)

		for {
			select {
			case <-tech:
				cancel()

				// Отправялем все имеющиеся значения
				for s, val := range mCandles {
					out <- *mCandles[s]

					// Запись в файл //
					if err := w.Write(toRecord(*val)); err != nil {
						return
					}
				}

				// Ждем завершения формирования 2-минутных свеч.

				<-sync
				close(out)
				return
			case price := <-prices:
				// Ввод только начался

				switch mCandles[price.Ticker].TS {
				case nullTime:
					initCandle(mCandles[price.Ticker], price, domain.CandlePeriod1m)
				case price.TS:
					updateCandle(mCandles[price.Ticker], price)
				default:
					// Настал новый период
					out <- *mCandles[price.Ticker]

					// Запись в файл //
					if err := w.Write(toRecord(*mCandles[price.Ticker])); err != nil {
						return
					}

					initCandle(mCandles[price.Ticker], price, domain.CandlePeriod1m)
				}
			}
		}
	}()
	return out
}

func twoMin(ctx context.Context, in <-chan domain.Candle, cancel context.CancelFunc, sync1, sync2 chan struct{}) <-chan domain.Candle {
	out := make(chan domain.Candle)

	// Хранилище для 1-минутых свечей.
	// Для формирования 2-минутной нужно 2 1-минутные
	mOneMinCandles := make(map[string][]domain.Candle)

	// Мапа хранит 2-х минутные свечи для каждой комапнии.
	mTwoMinCandle := make(map[string]*domain.Candle)
	initMapOfCandle(mTwoMinCandle, domain.CandlePeriod2m)

	go func() {
		f, _ := os.Create("candles_2m.csv")
		defer f.Close()

		w := csv.NewWriter(f)
		defer w.Flush()

		for {
			select {
			case <-ctx.Done():
				cancel()
				// Считаем все свечи из потока 1-минутных
				for s, val := range mOneMinCandles {
					oneMinCandle := <-in
					mOneMinCandles[s] = append(val, oneMinCandle)
				}
				// Формируем 2-х минутные.
				for s, val := range mTwoMinCandle {
					mTwoMinCandle[s] = buildCandle(mOneMinCandles[s], domain.CandlePeriod2m)
					out <- *mTwoMinCandle[s]
					// Запись в файл
					if err := w.Write(toRecord(*val)); err != nil {
						return
					}
				}

				// Ждем завершения формирования 10-минутных свеч.
				<-sync2

				close(out)

				// Делаем синхронизация с функций формирования 1-минутных свеч.
				sync1 <- struct{}{}
				return
			case candle := <-in:

				// Добавляем в хранилище
				mOneMinCandles[candle.Ticker] = append(mOneMinCandles[candle.Ticker], candle)

				// Если собрали 2 1-минутные свечки, формируем 2-минутную,
				// отправляем по пайплайну на формирование 10-минутной.
				if len(mOneMinCandles[candle.Ticker]) == 2 {
					mTwoMinCandle[candle.Ticker] = buildCandle(mOneMinCandles[candle.Ticker], domain.CandlePeriod2m)
					out <- *mTwoMinCandle[candle.Ticker]

					// Запись в файл
					if err := w.Write(toRecord(*mTwoMinCandle[candle.Ticker])); err != nil {
						return
					}
					mOneMinCandles[candle.Ticker] = nil
				}
			}
		}
	}()
	return out
}

func tenMin(ctx context.Context, in <-chan domain.Candle, sync chan struct{}) {
	// Для формирования 10-минутной свечи нужно 5 2-минутных.
	mTwoMinCandles := make(map[string][]domain.Candle)

	// Мапа хранит 10-минутные свечи для каждой компании.
	mTenMinCandle := make(map[string]*domain.Candle)
	initMapOfCandle(mTenMinCandle, domain.CandlePeriod10m)

	go func() {
		f, _ := os.Create("candles_10m.csv")
		defer f.Close()

		w := csv.NewWriter(f)
		defer w.Flush()
		for {
			select {
			case <-ctx.Done():
				// Считаем все свечи из потока 2-минутных
				for s, val := range mTwoMinCandles {
					TwoMinCandle := <-in
					mTwoMinCandles[s] = append(val, TwoMinCandle)
				}

				// Формируем 10-минутные свечи.

				for s, val := range mTenMinCandle {
					mTenMinCandle[s] = buildCandle(mTwoMinCandles[s], domain.CandlePeriod10m)

					// Запись в файл
					if err := w.Write(toRecord(*val)); err != nil {
						return
					}
				}

				// Синхронизируемся с функцией формирования 2-минутных свечей.
				sync <- struct{}{}
				return
			case candle := <-in:

				// Добавляем в хранилище.
				mTwoMinCandles[candle.Ticker] = append(mTwoMinCandles[candle.Ticker], candle)

				// Если собрали 5 2-минутные свечки, формируем 10-минутную.
				if len(mTwoMinCandles[candle.Ticker]) == 5 {
					mTenMinCandle[candle.Ticker] = buildCandle(mTwoMinCandles[candle.Ticker], domain.CandlePeriod10m)

					// Запись в файл
					if err := w.Write(toRecord(*mTenMinCandle[candle.Ticker])); err != nil {
						return
					}
					mTwoMinCandles[candle.Ticker] = nil
				}
			}
		}
	}()
}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  15,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})
	// Канал из которого поступают цены
	pg.Prices(ctx)

	// Файл для записи 1-минутных свечей.
	f1, _ := os.Create("candles_1m.csv")
	defer f1.Close()
	w1 := csv.NewWriter(f1)
	defer w1.Flush()

	// Файл для записи 2-минутных свечей.
	f2, _ := os.Create("candles_2m.csv")
	defer f2.Close()
	w2 := csv.NewWriter(f2)
	defer w2.Flush()

	// Файл для записи 10-минутных свечей.
	f10, _ := os.Create("candles_10m.csv")
	defer f10.Close()
	w10 := csv.NewWriter(f10)
	defer w10.Flush()

	// канал для обратоки сигнала SIGINT
	tech := make(chan os.Signal, 1)
	signal.Notify(tech, syscall.SIGINT)

	// Запускаем конвеер

	// Ждем ^C
	<-tech

	// Ждем завершения пайплайна
	//<-outLust

	ctx1, cancelFunc1 := context.WithCancel(ctx)

	sync1 := make(chan struct{})
	sync2 := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	out1 := oneMin(ctx, pg, wg, cancelFunc, sync1)
	out2 := twoMin(ctx, out1, cancelFunc1, sync1, sync2)
	tenMin(ctx1, out2, sync2)
	wg.Wait()
}
