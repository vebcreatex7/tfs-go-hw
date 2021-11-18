package services

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/tfs-go-hw/course_project/internal/repository"
)

type Bot struct {
	conn       *websocket.Conn
	repo       repository.Repository
	done       context.Context
	publicKey  string
	privateKey string
	symbol     string
}

type BotService interface {
	SetSymbol(string)
}

func NewBotService(c *websocket.Conn, r repository.Repository, d context.Context, public string, private string) BotService {
	return &Bot{
		conn:       c,
		repo:       r,
		done:       d,
		publicKey:  public,
		privateKey: private,
	}
}

func (b *Bot) SetSymbol(s string) {
	b.symbol = s
}
