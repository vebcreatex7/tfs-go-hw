package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"golang.org/x/sync/errgroup"
)

//Exchange
type ExchangeMock struct{}

func (em *ExchangeMock) SetSymbol(s string) {

}

func (em *ExchangeMock) GetSymbol() string {
	return "pi_xbtusd"
}

func (em *ExchangeMock) SetPeriod(period domain.CandlePeriod) {

}

func (em *ExchangeMock) GetPeriod() domain.CandlePeriod {
	return domain.CandlePeriod1m
}

func (em *ExchangeMock) SetAmount(amount int) {

}

func (em *ExchangeMock) WSConnect() error {
	return nil
}

func (em *ExchangeMock) WSDisconnect() error {
	return nil
}

func (em ExchangeMock) CandlesFlow(eg *errgroup.Group, c context.Context) <-chan domain.Candle {
	candle := make(chan domain.Candle)
	eg.Go(func() error {
		defer func() {
			close(candle)
		}()
		candle <- domain.Candle{}
		<-c.Done()
		return nil
	})

	return candle
}

func (em *ExchangeMock) GetOHLC(s string, p domain.CandlePeriod, n int64) ([]domain.Candle, error) {
	candles := make([]domain.Candle, n)
	return candles, nil
}

func (em *ExchangeMock) GetOpenPositions() error {
	return nil
}

func (em *ExchangeMock) SendOrderMkt(side string) (domain.Order, error) {
	d := domain.SendStatus{
		Status:       "placed",
		ReceivedTime: "2021-01-02T18:04:05.000Z",
		OrderEvents:  make([]domain.OrderEvent, 1),
	}
	return domain.Order{
		SendStatus: d,
	}, nil
}

// Indicator
type IndicatorMock struct{}

func (im *IndicatorMock) Init(candles []domain.Candle) error {
	return nil
}

func (im *IndicatorMock) CandlesNeeded() int {
	return 5
}

func (im *IndicatorMock) Indicate(candle domain.Candle) domain.Action {
	return domain.Buy
}

func (im *IndicatorMock) Config(fast, slow, signal int, s rune) {}

func (im *IndicatorMock) ChangeSource(s rune) {}

//Repo
type RepositoryMock struct{}

func (rm *RepositoryMock) InsertOrder(ctx context.Context, order domain.RecordOrder) error {
	return nil
}

// TgBot
type TgBotMock struct{}

func (tm *TgBotMock) SendOrder(order domain.RecordOrder) error {
	return nil
}

func TestExchange_Testify(t *testing.T) {
	exchange := &ExchangeMock{}
	indicator := &IndicatorMock{}
	repo := &RepositoryMock{}
	tgbot := &TgBotMock{}
	bot := NewBotService(repo, tgbot, nil, exchange, indicator)

	assert.Equal(t, "pi_xbtusd", bot.GetSymbol())
	assert.Equal(t, domain.CandlePeriod1m, bot.GetPeriod())
	assert.Equal(t, nil, bot.WSConnect())
	assert.Equal(t, nil, bot.WSDisconnect())
}

func TestIndicator_Testify(t *testing.T) {
	exchange := &ExchangeMock{}
	indicator := &IndicatorMock{}
	repo := &RepositoryMock{}
	tgbot := &TgBotMock{}
	bot := NewBotService(repo, tgbot, nil, exchange, indicator)
	assert.Equal(t, nil, bot.initIndicator())
}

func TestPipeline_Testify(t *testing.T) {
	exchange := &ExchangeMock{}
	indicator := &IndicatorMock{}
	repo := &RepositoryMock{}
	tgbot := &TgBotMock{}
	bot := NewBotService(repo, tgbot, nil, exchange, indicator)
	done, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	eg, _ := errgroup.WithContext(done)

	//Pipeline
	candle := bot.exchange.CandlesFlow(eg, done)
	action := bot.Indicator(eg, candle)
	order := bot.Trading(eg, action)
	bot.Record(eg, order)

	assert.Equal(t, nil, eg.Wait())
}
