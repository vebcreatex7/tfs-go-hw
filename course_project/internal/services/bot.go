package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services/kraken"
)

type Bot struct {
	repo   repository.Repository
	kraken kraken.KrakenService
}

func NewBotService(r repository.Repository, k kraken.KrakenService) BotService {
	return &Bot{
		repo:   r,
		kraken: k,
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

func (b *Bot) Run(done context.Context, finished chan struct{}) {

	// Connecting to market
	err := b.kraken.WSConnect()
	if err != nil {
		log.Println(err)
		finished <- struct{}{}
		return
	}
	log.Println("Connected to market")
	/*
		// Subscribing to candle flow
		err = b.kraken.WSSubscribe()
		if err != nil {
			log.Println(err)
			finished <- struct{}{}
			return
		}
		log.Println("Subscribed to market")
	*/
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stop, channelFunc := context.WithCancel(done)

	candle := make(chan domain.Candle)

	// Ping
	go b.kraken.WriteHandler(wg, stop)

	// Reading candles from ws connect
	go b.kraken.ReadHandler(wg, stop, candle)

	for {
		select {
		case <-done.Done():
			channelFunc()
			wg.Wait()
			/*
				// Unsubscribing from candle flow
				err = b.kraken.WSUnsubscribe()
				if err != nil {
					log.Println(err)
				}
				log.Println("Unsubscribed from market")
			*/
			// Disconnecting from market
			err = b.kraken.WSDisconnect()
			if err != nil {
				log.Println(err)
				finished <- struct{}{}
				return
			}
			log.Println("Disconnected from market")

			close(candle)
			finished <- struct{}{}
			return
		case candle := <-candle:
			fmt.Println(candle)
		}
	}

}
