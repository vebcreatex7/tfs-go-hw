package services

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/tfs-go-hw/course_project/internal/repository"
	"github.com/tfs-go-hw/course_project/internal/services/kraken"
)

type Bot struct {
	repo   repository.Repository
	kraken kraken.KrakenService
}

func NewBotService(c *websocket.Conn, r repository.Repository, public string, private string) BotService {
	return &Bot{
		repo:   r,
		kraken: kraken.NewKraken(c, public, private),
	}
}

type BotService interface {
	SetSymbol(string)
	Run(context.Context) chan struct{}
	GetSymbol() string
}

func (b *Bot) SetSymbol(s string) {
	b.kraken.SetSymbol(s)
}

func (b *Bot) GetSymbol() string {
	return b.kraken.GetSymbol()
}

func (b *Bot) Run(done context.Context) chan struct{} {
	finish := make(chan struct{})
	return finish
}
