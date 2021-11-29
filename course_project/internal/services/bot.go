package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"golang.org/x/sync/errgroup"
)

var (
	// Time allowed to get candle for initialization indicator
	candleWait = 60 * time.Second
)

var ErrInitIndicator = errors.New("InitIndicatorError")

type Exchange interface {
	SetSymbol(symbol string)
	GetSymbol() string
	SetPeriod(period domain.CandlePeriod)
	GetPeriod() domain.CandlePeriod
	WSConnect() error
	WSDisconnect() error
	CandlesFlow(*errgroup.Group, context.Context) <-chan domain.Candle
	GetOHLC(s string, p domain.CandlePeriod, n int64) ([]domain.Candle, error)
	GetOpenPositions() error
	SendOrderMkt(side string) (domain.Order, error)
}

type Indicator interface {
	Init(candles []domain.Candle) error
	CandlesNeeded() int
	Indicate(candle domain.Candle) domain.Action
}

type Repository interface {
	InsertOrder(ctx context.Context, order domain.RecordOrder) error //postgres
}
type TgSender interface {
	SendOrder(order domain.RecordOrder) error
}

type Bot struct {
	repo      Repository
	tgBot     TgSender
	exchange  Exchange
	indicator Indicator
	logger    logrus.FieldLogger
}

func NewBotService(r Repository, t TgSender, l logrus.FieldLogger, k Exchange, m Indicator) *Bot {
	return &Bot{
		repo:      r,
		tgBot:     t,
		logger:    l,
		exchange:  k,
		indicator: m,
	}
}

func (b *Bot) SetSymbol(s string) {
	b.exchange.SetSymbol(s)
}

func (b *Bot) GetSymbol() string {
	return b.exchange.GetSymbol()
}

func (b *Bot) SetPeriod(s domain.CandlePeriod) {
	b.exchange.SetPeriod(s)
}

func (b *Bot) GetPeriod() domain.CandlePeriod {
	return b.exchange.GetPeriod()
}

func (b *Bot) WSConnect() error {
	return b.exchange.WSConnect()
}

func (b *Bot) WSDisconnect() error {
	return b.exchange.WSDisconnect()
}

func (b *Bot) initIndicator() error {
	wait := time.NewTicker(candleWait)
	for {
		select {
		case <-wait.C:
			return ErrInitIndicator
		default:
			candles, err := b.exchange.GetOHLC(b.exchange.GetSymbol(), b.exchange.GetPeriod(), int64(b.indicator.CandlesNeeded())+1)
			if err != nil {
				continue
			}
			// Remove current price
			candles = candles[:len(candles)-1]
			// Init indicator
			err = b.indicator.Init(candles)
			if err == nil {
				return err
			}

		}
	}
}

// Part of the pipline. Uses indicator to make decision
func (b *Bot) Indicator(eg *errgroup.Group, candle <-chan domain.Candle) <-chan domain.Action {
	action := make(chan domain.Action)

	eg.Go(func() error {
		defer func() {
			close(action)
		}()
		for c := range candle {
			action <- b.indicator.Indicate(c)
		}
		return nil
	})
	return action
}

// Part of the pipline. Makes orders on the exchange.
func (b *Bot) Trading(eg *errgroup.Group, action <-chan domain.Action) <-chan domain.RecordOrder {
	order := make(chan domain.RecordOrder)

	eg.Go(func() error {
		defer func() {
			close(order)
		}()
		for a := range action {
			if a == domain.Buy || a == domain.Sell {
				o, err := b.exchange.SendOrderMkt(string(a))
				if err != nil {
					return err
				}

				if o.Error != "" {
					return fmt.Errorf("errRecord: %s", o.Error)
				}
				if o.SendStatus.Status != "placed" {
					return fmt.Errorf("errStatus: %s", o.SendStatus.Status)
				}
				record, err := domain.NewRecordOrder(&o)
				if err != nil {
					return err
				}

				order <- *record
			}
		}
		return nil
	})

	return order
}

// Last part of pipline. Records orders to the postgres DB and sends to the tgBot.
func (b *Bot) Record(eg *errgroup.Group, order <-chan domain.RecordOrder) {
	eg.Go(func() error {
		for o := range order {
			err := b.repo.InsertOrder(context.Background(), o)
			if err != nil {
				return err
			}
			err = b.tgBot.SendOrder(o)
			if err != nil {
				return err
			}

		}
		return nil
	})
}

// Starts the pipline
func (b *Bot) Run(ctx context.Context, finished chan struct{}) {

	defer func() {
		finished <- struct{}{}
	}()

	// Get open positions to work with
	err := b.exchange.GetOpenPositions()
	if err != nil {
		b.logger.Println(err)
		return
	}

	// Init indicator
	err = b.initIndicator()
	if err != nil {
		b.logger.Println(err)
		return
	}

	for {
		// Connecting to the exchange
		err = b.exchange.WSConnect()
		if err != nil {
			b.logger.Println(err)
			return
		}
		b.logger.Println("Connected to the exchange")

		eg, errDone := errgroup.WithContext(ctx)
		done, channelFunc := context.WithCancel(ctx)

		// Pipline
		candle := b.exchange.CandlesFlow(eg, done)
		action := b.Indicator(eg, candle)
		order := b.Trading(eg, action)
		b.Record(eg, order)

		select {
		case <-ctx.Done():
			channelFunc()
			if err = eg.Wait(); err == nil {
				b.logger.Println("Pipeline stoped successfully")
			}
			err = b.exchange.WSDisconnect()
			if err != nil {
				b.logger.Println(err)
				return
			}
			b.logger.Println("Disconnected from the exchange")
			return
		case <-errDone.Done():
			channelFunc()
			err = eg.Wait()
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				b.logger.Println(err)
				b.logger.Println("Reconnecting...")
				continue
			} else {
				b.logger.Println(err)
				b.logger.Println("Pipeline stoped with error")
				err = b.exchange.WSDisconnect()
				if err != nil {
					b.logger.Println(err)
				}
				return
			}
		}
	}

}
