package services

import (
	"context"
	"log"

	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services/indicators"
	"github.com/tfs-go-hw/course_project/internal/services/kraken"
	"golang.org/x/sync/errgroup"
)

type Bot struct {
	repo   repository.Repository
	kraken kraken.KrakenService
	macd   indicators.MacdService
}

func NewBotService(r repository.Repository, k kraken.KrakenService, m indicators.MacdService) BotService {
	return &Bot{
		repo:   r,
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

func (b *Bot) Run(ctx context.Context, finished chan struct{}) {

	defer func() {
		finished <- struct{}{}
	}()

	// Init indicator
	candles, err := b.kraken.GetOHLC(b.kraken.GetSymbol(), b.kraken.GetPeriod(), int64(b.macd.CandlesNeeded())+1)
	if err != nil {
		log.Println(err)
		return
	}

	// Remove current price
	candles = candles[:len(candles)-1]

	err = b.macd.InitMacd(candles)
	if err != nil {
		log.Println(err)
		return
	}

	// Connecting to market
	err = b.kraken.WSConnect()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Connected to market")

	eg, errDone := errgroup.WithContext(ctx)
	done, channelFunc := context.WithCancel(ctx)

	// Pipline
	candle := b.kraken.CandlesFlow(eg, done)
	b.macd.Serve(eg, candle)

	select {
	case <-ctx.Done():
		channelFunc()
		if err = eg.Wait(); err == nil {
			log.Println("Pipeline stoped successfully")
		}
		err = b.kraken.WSDisconnect()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Disconnected from market")
		return
	case <-errDone.Done():
		channelFunc()
		err = eg.Wait()
		if err != nil {
			log.Println(err)
			log.Println("Pipeline stoped unsuccessfully")
		}
		err = b.kraken.WSDisconnect()
		if err != nil {
			log.Println(err)
		}
		return
	}
}
