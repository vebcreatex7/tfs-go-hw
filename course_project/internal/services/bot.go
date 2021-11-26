package services

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (b *Bot) Run(ctx context.Context, finished chan struct{}) {

	defer func() {
		finished <- struct{}{}
	}()

	// Get open positions to work with
	err := b.kraken.GetOpenPositions()
	if err != nil {
		b.logger.Println(err)
		//log.Println(err)
		return
	}

	// Init indicator
	err = b.initIndicator()
	if err != nil {
		b.logger.Println(err)
		//log.Println(err)
		return
	}

	// Connecting to market
	err = b.kraken.WSConnect()
	if err != nil {
		b.logger.Println(err)
		//log.Println(err)
		return
	}
	b.logger.Println("Connected to market")
	//log.Println("Connected to market")

	eg, errDone := errgroup.WithContext(ctx)
	done, channelFunc := context.WithCancel(ctx)

	// Pipline
	candle := b.kraken.CandlesFlow(eg, done)
	action := b.macd.Serve(eg, candle)
	order := b.kraken.Trade(eg, action)
	eg.Go(func() error {
		for o := range order {
			fmt.Printf("%#v", o)
		}
		return nil
	})

	select {
	case <-ctx.Done():
		channelFunc()
		if err = eg.Wait(); err == nil {
			b.logger.Println("Pipeline stoped successfully")
			//log.Println("Pipeline stoped successfully")
		}
		err = b.kraken.WSDisconnect()
		if err != nil {
			b.logger.Println(err)
			//log.Println(err)
			return
		}
		b.logger.Println("Disconnected from market")
		//log.Println("Disconnected from market")
		return
	case <-errDone.Done():
		channelFunc()
		err = eg.Wait()
		if err != nil {
			b.logger.Println(err)
			//log.Println(err)
			b.logger.Println("Pipeline stoped with error")
			//log.Println("Pipeline stoped with error")
		}
		err = b.kraken.WSDisconnect()
		if err != nil {
			b.logger.Println(err)
			//log.Println(err)
		}
		return
	}
}
