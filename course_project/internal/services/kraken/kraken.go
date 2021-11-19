package kraken

import "github.com/gorilla/websocket"

type Kraken struct {
	conn       *websocket.Conn
	publicKey  string
	privateKey string
	symbol     string
}

type KrakenService interface {
	SetSymbol(string)
	GetSymbol() string
}

func NewKraken(c *websocket.Conn, public string, private string) KrakenService {
	return &Kraken{
		conn:       c,
		publicKey:  public,
		privateKey: private,
	}
}

func (k *Kraken) SetSymbol(s string) {
	k.symbol = s
}

func (k *Kraken) GetSymbol() string {
	return k.symbol
}
