package services

import (
	"context"
	"errors"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services/indicators"
	"github.com/tfs-go-hw/course_project/internal/services/kraken"
	"golang.org/x/sync/errgroup"
)

var (
	// Time allowed to get candle for initialization indicator
	candleWait = 60 * time.Second
)

var ErrInitIndicator = errors.New("InitIndicatorError")

type Bot struct {
	repo   repository.Repository
	kraken kraken.KrakenService
	macd   indicators.MacdService
	logger logrus.FieldLogger
}

func NewBotService(r repository.Repository, l logrus.FieldLogger, k kraken.KrakenService, m indicators.MacdService) BotService {
	return &Bot{
		repo:   r,
		logger: l,
		kraken: k,
		macd:   m,
	}
}

type BotService interface {
	Run(context.Context, chan struct{})
	SetSymbol(string)
	GetSymbol() string
	SetPeriod(domain.CandlePeriod)
	GetPeriod() domain.CandlePeriod
	WSConnect() error
	WSDisconnect() error
}

func (b *Bot) SetSymbol(s string) {
	b.kraken.SetSymbol(s)
}

func (b *Bot) GetSymbol() string {
	return b.kraken.GetSymbol()
}

func (b *Bot) SetPeriod(s domain.CandlePeriod) {
	b.kraken.SetPeriod(s)
}

func (b *Bot) GetPeriod() domain.CandlePeriod {
	return b.kraken.GetPeriod()
}

func (b *Bot) WSConnect() error {
	return b.kraken.WSConnect()
}

func (b *Bot) WSDisconnect() error {
	return b.kraken.WSDisconnect()
}

func (b *Bot) initIndicator() error {
	wait := time.NewTicker(candleWait)
	for {
		select {
		case <-wait.C:
			return ErrInitIndicator
		default:
			candles, err := b.kraken.GetOHLC(b.kraken.GetSymbol(), b.kraken.GetPeriod(), int64(b.macd.CandlesNeeded())+1)
			if err != nil {
				continue
			}
			// Remove current price
			candles = candles[:len(candles)-1]
			// Init indicator
			err = b.macd.InitMacd(candles)
			if err == nil {
				return err
			}

		}
	}
}

// Last part of pipline, records orders to the postgres DB and sends to the tgBot
func (b *Bot) Record(eg *errgroup.Group, order <-chan domain.RecordOrder) {
	eg.Go(func() error {
		for o := range order {
			err := b.repo.InsertOrder(context.TODO(), o)
			if err != nil {
				return err
			}
			err = b.repo.SendOrder(o)
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
	err := b.kraken.GetOpenPositions()
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
		err = b.kraken.WSConnect()
		if err != nil {
			b.logger.Println(err)
			return
		}
		b.logger.Println("Connected to the exchange")

		eg, errDone := errgroup.WithContext(ctx)
		done, channelFunc := context.WithCancel(ctx)

		// Pipline
		candle := b.kraken.CandlesFlow(eg, done)
		action := b.macd.Serve(eg, candle)
		order := b.kraken.Trade(eg, action)
		b.Record(eg, order)

		select {
		case <-ctx.Done():
			channelFunc()
			if err = eg.Wait(); err == nil {
				b.logger.Println("Pipeline stoped successfully")
			}
			err = b.kraken.WSDisconnect()
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
				err = b.kraken.WSDisconnect()
				if err != nil {
					b.logger.Println(err)
				}
				return
			}
		}
	}

}
