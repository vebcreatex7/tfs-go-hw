package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/tfs-go-hw/course_project/internal/domain"
)

type BotService interface {
	Run(<-chan struct{}, chan<- struct{})
	SetSymbol(string)
	GetSymbol() string
	SetPeriod(domain.CandlePeriod)
	GetPeriod() domain.CandlePeriod
	SetAmount(amount int)
	ChangeSourceIndicator(s rune)
	ConfigurateIndicator(fast, slow, signal int, s rune)
}

type Bot struct {
	start     chan struct{}
	stop      chan struct{}
	done      context.Context
	service   BotService
	logger    logrus.FieldLogger
	isRunning bool
}

func NewBot(d context.Context, s BotService, l logrus.FieldLogger) *Bot {
	return &Bot{
		start:   make(chan struct{}),
		stop:    make(chan struct{}),
		done:    d,
		service: s,
		logger:  l,
	}
}

func (b *Bot) Run(wg *sync.WaitGroup) {

	go func() {

		// Channel to stop the bot
		stopService := make(chan struct{})

		// channel to sync with the bot
		serviceStoped := make(chan struct{})

		for {

			select {
			// Stop the app before running the bot
			case <-b.done.Done():
				if b.isRunning {
					stopService <- struct{}{}
					<-serviceStoped
					close(serviceStoped)
					close(stopService)
					b.isRunning = false
				}
				b.logger.Println("app is stopped")
				close(b.start)
				close(b.stop)
				wg.Done()
				return

			// Start the bot
			case <-b.start:
				b.isRunning = true
				b.logger.Println("bot is running")
				go b.service.Run(stopService, serviceStoped)

			// Stop the bot
			case <-b.stop:
				stopService <- struct{}{}
				<-serviceStoped
				b.logger.Println("bot is stoped")
				b.isRunning = false

			case <-serviceStoped:
				<-stopService
				b.isRunning = false
				b.logger.Println("Internal bot error")

			}
		}
	}()

}

func (b *Bot) Start(w http.ResponseWriter, r *http.Request) {
	if b.service.GetSymbol() == "" || b.service.GetPeriod() == "" || b.isRunning {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b.start <- struct{}{}
}

func (b *Bot) Stop(w http.ResponseWriter, r *http.Request) {
	if !b.isRunning {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b.stop <- struct{}{}
}

func (b *Bot) ConfigurateIndicator(w http.ResponseWriter, r *http.Request) {
	if b.isRunning {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("It isn't possible to change this params during the work"))
		return
	}
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}
	defer r.Body.Close()

	buf := &domain.Indicator{}
	err = json.Unmarshal(d, buf)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}
	if !buf.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}

	sourceRune := []rune(buf.Source)[0]

	b.service.ConfigurateIndicator(buf.Fast, buf.Slow, buf.Signal, sourceRune)
	_, _ = w.Write([]byte("Ok"))
}

func (b *Bot) ChangeSourceIndicator(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}
	defer r.Body.Close()
	buf := &domain.Source{}
	err = json.Unmarshal(d, buf)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}
	if !buf.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}
	sourceRune := []rune(buf.Source)[0]
	b.service.ChangeSourceIndicator(sourceRune)
	_, _ = w.Write([]byte("Ok"))

}

func (b *Bot) ConfigurateExchange(w http.ResponseWriter, r *http.Request) {
	if b.isRunning {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("It isn't possible to change this params during the work"))
		return
	}
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}
	defer r.Body.Close()

	buf := domain.Exchange{}

	err = json.Unmarshal(d, &buf)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}
	if !buf.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Body"))
		return
	}

	b.service.SetSymbol(buf.Symbol.Symbol)
	b.service.SetPeriod(buf.Period.Period)
	b.service.SetAmount(buf.Amount.Amount)

}

func (b *Bot) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/bot", func(r chi.Router) {
		r.Post("/start", b.Start)
		r.Post("/stop", b.Stop)
		r.Post("/exchange/config", b.ConfigurateExchange)
		r.Post("/indicator/config", b.ConfigurateIndicator)
		r.Post("/indicator/change_source", b.ChangeSourceIndicator)

	})

	return r
}
