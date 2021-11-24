package kraken

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tfs-go-hw/course_project/internal/domain"
	"golang.org/x/sync/errgroup"
)

const (

	// Time allowed to subscibe to the candle stream
	subscibeWait = 30 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 30 * time.Second

	// Send pings to peer with this period.
	pingPeriod = (pongWait * 9) / 10
)

const urlWs = "wss://demo-futures.kraken.com/ws/v1?chart"

var (
	ErrBadConnect = errors.New("kraken: bad subscribe")
)

type Kraken struct {
	conn       *websocket.Conn
	publicKey  string
	privateKey string
	symbol     string
	period     domain.CandlePeriod
}

func NewKraken(public string, private string) KrakenService {
	return &Kraken{
		publicKey:  public,
		privateKey: private,
	}
}

type KrakenService interface {
	SetSymbol(string)
	GetSymbol() string
	SetPeriod(domain.CandlePeriod)
	GetPeriod() domain.CandlePeriod
	WSConnect() error
	WSDisconnect() error
	CandlesFlow(*errgroup.Group, context.Context) <-chan domain.Candle
}

func (k *Kraken) SetSymbol(s string) {
	k.symbol = s
}

func (k *Kraken) GetSymbol() string {
	return k.symbol
}

func (k *Kraken) SetPeriod(s domain.CandlePeriod) {
	k.period = s
}

func (k *Kraken) GetPeriod() domain.CandlePeriod {
	return k.period
}

func (k *Kraken) WSConnect() error {
	wait := time.NewTicker(subscibeWait)
	var err error
	for {
		select {
		case <-wait.C:
			k.conn = nil
			return ErrBadConnect
		default:
			k.conn, _, err = websocket.DefaultDialer.Dial(urlWs, nil)
			if err == nil {
				err = k.conn.SetReadDeadline(time.Now().Add(pongWait))
				if err != nil {
					log.Panicln(err)
					return err
				}
				k.conn.SetPongHandler(func(string) error { k.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

				return nil
			}
		}
	}
}

func (k *Kraken) WSDisconnect() error {
	err := k.conn.Close()
	k.conn = nil
	if err != nil {
		return err
	}

	return nil
}

func (k *Kraken) CandlesFlow(eg *errgroup.Group, ctx context.Context) <-chan domain.Candle {

	c := make(chan domain.Candle)

	// Gorutine subscribes then pings
	eg.Go(func() error {
		ping := time.NewTicker(pingPeriod)
		defer func() {
			ping.Stop()
		}()

		// Subscribe
		event := domain.NewEvent("subscribe", string(k.period), k.symbol)
		err := k.conn.WriteJSON(event)
		if err != nil {
			//log.Println(err)
			return err
		}

		// Ping
		for {
			select {
			case <-ping.C:
				err = k.conn.WriteMessage(websocket.PingMessage, nil)
				if err != nil {
					//log.Println(err)
					return err
				}
			case <-ctx.Done():
				stop := event
				stop.Event = "unsubscribe"
				err = k.conn.WriteJSON(stop)
				if err != nil {
					//log.Panicln(err)
					return err
				}
				return nil
			}
		}

	})

	// Gorutine reads tnen sends candles over a chan
	eg.Go(func() error {
		defer func() {
			close(c)
		}()

		// Read version
		version := &domain.Event{}
		err := k.conn.ReadJSON(version)
		if err != nil {
			//log.Panicln(err)
			return err
		}

		// Subscribed
		var event domain.Event
		err = k.conn.ReadJSON(&event)
		if err != nil {
			//log.Println(err)
			return err
		}

		// Candle snapshot
		candle := &domain.Candle{}
		tmp := domain.CandleSubscribe{}
		err = k.conn.ReadJSON(&tmp)
		if err != nil {
			//log.Println(err)
			return err
		}

		candle.BuildCandle(tmp)
		c <- *candle

		// Candle flow
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				err := k.conn.ReadJSON(&tmp)
				if err != nil {
					return err
				}
				if tmp.C.Time == candle.Time {
					candle.Close = tmp.C.Close
					if tmp.C.Low < candle.Low {
						candle.Low = tmp.C.Low
					}
					if tmp.C.High > candle.High {
						candle.High = tmp.C.High
					}
					candle.Volume += tmp.C.Volume
				} else {
					c <- *candle
					candle.BuildCandle(tmp)
				}
			}
		}

	})
	return c
}

func (k *Kraken) genAuth(postData string, endPoint string) (string, error) {
	sha := sha256.New()
	concat := postData + endPoint
	sha.Write([]byte(concat))

	apiDecode, err := base64.StdEncoding.DecodeString(k.privateKey)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha512.New, apiDecode)
	h.Write(sha.Sum(nil))

	result := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return result, nil
}
