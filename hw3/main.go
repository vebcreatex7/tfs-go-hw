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

func toRecord(c *domain.Candle) []string {
	record := make([]string, 0)
	record = append(record, c.Ticker)
	record = append(record, c.TS.String())
	record = append(record, fmt.Sprintf("%f", c.Open))
	record = append(record, fmt.Sprintf("%f", c.High))
	record = append(record, fmt.Sprintf("%f", c.Low))
	record = append(record, fmt.Sprintf("%f", c.Close))

	return record

}

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

func oneMin(pg *generator.PricesGenerator, wg *sync.WaitGroup, ctx context.Context, cancel context.CancelFunc, sync chan struct{}) <-chan domain.Candle {

	// По каналу идут сгенерированные цены.
	prices := pg.Prices(ctx)

	// Канал, по которому передаются 1-минутные свечи.
	out := make(chan domain.Candle)

	// мапа хранит свечи для каждой комапании.
	mCandles := make(map[string]*domain.Candle)
	for _, ticker := range tickers {
		mCandles[ticker] = &domain.Candle{Period: domain.CandlePeriod1m}
	}

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
				for s, _ := range mCandles {
					out <- *mCandles[s]

					// Запись в файл //
					w.Write(toRecord(mCandles[s]))
				}

				// Ждем завершения формирования 2-минутных свеч.

				<-sync
				close(out)
				return
			case price := <-prices:
				// Ввод только начался
				if mCandles[price.Ticker].TS == nullTime {
					initCandle(mCandles[price.Ticker], price, domain.CandlePeriod1m)
				} else {

					// Настал новый период
					if mCandles[price.Ticker].TS != price.TS {
						out <- *mCandles[price.Ticker]

						// Запись в файл //
						w.Write(toRecord(mCandles[price.Ticker]))

						initCandle(mCandles[price.Ticker], price, domain.CandlePeriod1m)
					} else {
						updateCandle(mCandles[price.Ticker], price)
					}
				}

			}
		}
	}()
	return out
}

func twoMin(in <-chan domain.Candle, ctx context.Context, sync1, sync2 chan struct{}) <-chan domain.Candle {
	out := make(chan domain.Candle)

	// Хранилище для 1-минутых свечей.
	// Для формирования 2-минутной нужно 2 1-минутные
	mOneMinCandles := make(map[string][]domain.Candle)

	// Мапа хранит 2-х минутные свечи для каждой комапнии.
	mTwoMinCandle := make(map[string]*domain.Candle)
	for _, ticker := range tickers {
		mTwoMinCandle[ticker] = &domain.Candle{Period: domain.CandlePeriod2m}
	}

	go func() {
		f, _ := os.Create("candles_2m.csv")
		defer f.Close()

		w := csv.NewWriter(f)
		defer w.Flush()

		for {
			select {
			case <-ctx.Done():
				// Считаем все свечи из потока 1-минутных
				for s, _ := range mOneMinCandles {
					oneMinCandle := <-in
					mOneMinCandles[s] = append(mOneMinCandles[s], oneMinCandle)
				}
				// Формируем 2-х минутные.
				for s, _ := range mTwoMinCandle {
					mTwoMinCandle[s] = buildCandle(mOneMinCandles[s], domain.CandlePeriod2m)
					out <- *mTwoMinCandle[s]
					// Запись в файл
					w.Write(toRecord(mTwoMinCandle[s]))
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
					w.Write(toRecord(mTwoMinCandle[candle.Ticker]))
					mOneMinCandles[candle.Ticker] = nil
				}
			}
		}
	}()
	return out
}

func tenMin(in <-chan domain.Candle, ctx context.Context, sync chan struct{}) {

	// Для формирования 10-минутной свечи нужно 5 2-минутных.
	mTwoMinCandles := make(map[string][]domain.Candle)

	// Мапа хранит 10-минутные свечи для каждой компании.
	mTenMinCandle := make(map[string]*domain.Candle)
	for _, ticker := range tickers {
		mTenMinCandle[ticker] = &domain.Candle{Period: domain.CandlePeriod10m}
	}

	go func() {

		f, _ := os.Create("candles_10m.csv")
		defer f.Close()

		w := csv.NewWriter(f)
		defer w.Flush()
		for {
			select {
			case <-ctx.Done():
				// Считаем все свечи из потока 2-минутных
				for s, _ := range mTwoMinCandles {
					TwoMinCandle := <-in
					mTwoMinCandles[s] = append(mTwoMinCandles[s], TwoMinCandle)
				}

				// Формируем 10-минутные свечи.

				for s, _ := range mTenMinCandle {
					mTenMinCandle[s] = buildCandle(mTwoMinCandles[s], domain.CandlePeriod10m)

					//Запись в файл
					w.Write(toRecord(mTenMinCandle[s]))
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

					//Запись в файл
					w.Write(toRecord(mTenMinCandle[candle.Ticker]))
					mTwoMinCandles[candle.Ticker] = nil
				}
			}
		}
	}()
}

func main() {

	//ctx, cancelFunc := context.WithCancel(context.Background())

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  15,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	//prices := pg.Prices(ctx)

	ctx, cancelFunc := context.WithCancel(context.Background())
	ctx1, _ := context.WithCancel(ctx)

	sync1 := make(chan struct{})
	sync2 := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	out1 := oneMin(pg, wg, ctx, cancelFunc, sync1)
	out2 := twoMin(out1, ctx, sync1, sync2)
	tenMin(out2, ctx1, sync2)
	wg.Wait()
}