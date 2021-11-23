package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/services"
)

type Bot struct {
	start   chan struct{}
	stop    chan struct{}
	done    context.Context
	service services.BotService
}

func NewBot(d context.Context, s services.BotService) *Bot {
	return &Bot{
		start:   make(chan struct{}),
		stop:    make(chan struct{}),
		done:    d,
		service: s,
	}
}

func (b *Bot) Run(wg *sync.WaitGroup) {

	go func() {

		// channel to sync with the bot
		serviceStoped := make(chan struct{})

		for {
			// context to stop the bot
			serviceDone, cancelFunc := context.WithCancel(b.done)

			select {
			// Stop the app before running the bot
			case <-b.done.Done():
				cancelFunc()
				wg.Done()
				log.Println("app is stopped")
				return

			// Start the bot
			case <-b.start:
				log.Println("bot is running")
				go b.service.Run(serviceDone, serviceStoped)
			}

			select {
			// Stop the app and then the bot
			case <-b.done.Done():
				cancelFunc()
				<-serviceStoped
				wg.Done()
				log.Println("app and bot are stopped")
				return

			// Stop the bot
			case <-b.stop:
				cancelFunc()
				<-serviceStoped
				log.Println("bot is stoped")

			// Internal bot error
			case <-serviceStoped:
				log.Println("Internal bot error")
			}

		}
	}()

}

func (b *Bot) Start(w http.ResponseWriter, r *http.Request) {
	if b.service.GetSymbol() == "" || b.service.GetPeriod() == "" {
		//w.Write([]byte("not enough parameters"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	b.start <- struct{}{}
}

func (b *Bot) Stop(w http.ResponseWriter, r *http.Request) {
	b.stop <- struct{}{}
}

func (b *Bot) SetSymbol(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	buf := &domain.Symbol{}

	err = json.Unmarshal(d, buf)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !buf.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b.service.SetSymbol(buf.Symbol)

}

func (b *Bot) SetPeriod(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	buf := &domain.Period{}

	err = json.Unmarshal(d, buf)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !buf.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b.service.SetPeriod(buf.Period)
}

func (b *Bot) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/bot", func(r chi.Router) {
		r.Post("/start", b.Start)
		r.Post("/stop", b.Stop)
		r.Post("/set_symbol", b.SetSymbol)
		r.Post("/set_period", b.SetPeriod)
	})
	return r
}
