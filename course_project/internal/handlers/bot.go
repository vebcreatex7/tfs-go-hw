package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services"
)

type Bot struct {
	start   chan struct{}
	stop    chan struct{}
	done    context.Context
	service services.BotService
}

func NewBot(c *websocket.Conn, r repository.Repository, d context.Context, public string, private string) *Bot {
	return &Bot{
		start:   make(chan struct{}),
		stop:    make(chan struct{}),
		done:    d,
		service: services.NewBotService(c, r, public, private),
	}
}

func (b *Bot) Run(wg *sync.WaitGroup) {

	go func() {

		// channel to sync with the bot
		serviceStoped := make(chan struct{})

		// context to stop the bot
		doneService, cancelFunc := context.WithCancel(b.done)

		defer func() {
			cancelFunc()
			<-serviceStoped
			wg.Done()
		}()

		for {

			select {
			// Stop the app before running the bot
			case <-b.done.Done():
				return

			// Start the bot
			case <-b.start:
				go func() { serviceStoped = b.service.Run(doneService) }()
			}

			select {
			// Stop the app and then the bot
			case <-b.done.Done():
				return

			// Stop the bot
			case <-b.stop:
				cancelFunc()
				<-serviceStoped

			// Internal bot error
			case <-serviceStoped:

			}

		}
	}()

}

func (b *Bot) Start(w http.ResponseWriter, r *http.Request) {
	if b.service.GetSymbol() == "" {
		w.Write([]byte("Symbol is not set"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	b.start <- struct{}{}
}

func (b *Bot) Stop(w http.ResponseWriter, r *http.Request) {

	b.stop <- struct{}{}
}

func (b *Bot) Symbol(w http.ResponseWriter, r *http.Request) {

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

func (b *Bot) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/bot", func(r chi.Router) {
		r.Post("/start", b.Start)
		r.Post("/stop", b.Stop)
		r.Post("/set_symbol", b.Symbol)
	})
	return r
}
