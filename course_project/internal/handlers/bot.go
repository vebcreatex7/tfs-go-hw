package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services"
)

type Bot struct {
	start   chan struct{}
	stop    chan struct{}
	symbol  []string
	done    context.Context
	service services.BotService
}

func NewBot(c *websocket.Conn, r repository.Repository, d context.Context, public string, private string) *Bot {
	return &Bot{
		start:   make(chan struct{}),
		stop:    make(chan struct{}),
		symbol:  make([]string, 0),
		done:    d,
		service: services.NewBotService(c, r, d, public, private),
	}
}

func (b *Bot) Run() {

	<-b.start
	//go b.service.MakeMoney()
	<-b.stop
}

func (b *Bot) Start(w http.ResponseWriter, r *http.Request) {
	if len(b.symbol) == 0 {
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

	b.symbol = buf.GetValid()
	b.service.SetSymbol(b.symbol[0])

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
