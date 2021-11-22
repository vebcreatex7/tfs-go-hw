package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services/kraken"
)

type Bot struct {
	repo   repository.Repository
	kraken kraken.KrakenService
}

func NewBotService(r repository.Repository, public string, private string) BotService {
	return &Bot{
		repo:   r,
		kraken: kraken.NewKraken(public, private),
	}
}

type BotService interface {
	Run(context.Context, chan struct{})
	SetSymbol(string)
	GetSymbol() string
	SetPeriod(domain.CandlePeriod)
	GetPeriod() domain.CandlePeriod
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

func (b *Bot) Run(done context.Context, finished chan struct{}) {

	err := b.kraken.Subscribe()
	if err != nil {
		finished <- struct{}{}
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	stop, channelFunc := context.WithCancel(done)

	c := make(chan domain.Candle)

	go b.kraken.WriteHandler(wg, stop)
	go b.kraken.ReadHandler(wg, stop, c)
	for {
		select {
		case <-done.Done():
			channelFunc()
			b.kraken.CloseConnection()
			wg.Wait()
			finished <- struct{}{}
			return
		case candle := <-c:
			fmt.Println(candle)
		}
	}

}
