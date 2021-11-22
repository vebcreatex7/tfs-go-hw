package kraken

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tfs-go-hw/course_project/internal/domain"
)

const (

	// Time allowed to subscibe to the candle stream
	subscibeWait = 120 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period.
	pingPeriod = (pongWait * 9) / 10

	// Time allowed to connect to the market
	connWait = 30 * time.Second
)

const urlWs = "wss://demo-futures.kraken.com/ws/v1?chart"

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
	Subscribe() error
	CloseConnection() error
	ReadHandler(*sync.WaitGroup, context.Context, chan domain.Candle)
	WriteHandler(*sync.WaitGroup, context.Context)
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

func SetConnection(url string) (*websocket.Conn, error) {
	wait := time.NewTicker(connWait)
	var conn *websocket.Conn
	var err error
	for {
		select {
		case <-wait.C:
			return nil, fmt.Errorf("%s", "can't establish a connection")
		default:
			conn, _, err = websocket.DefaultDialer.Dial(url, nil)
			if err == nil {
				return conn, nil
			}
		}
	}
}

func (k *Kraken) CloseConnection() error {
	return k.conn.Close()
}

func (k *Kraken) Subscribe() error {
	var err error
	k.conn, err = SetConnection(urlWs)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kraken) WriteHandler(wg *sync.WaitGroup, done context.Context) {
	ping := time.NewTicker(pingPeriod)

	defer func() {
		ping.Stop()
		wg.Done()
		k.conn.Close()
	}()

	event := domain.NewEvent("subscribe", string(k.period), k.symbol)
	k.conn.WriteJSON(event)

	for {
		select {
		case <-ping.C:
			if err := k.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		case <-done.Done():
			stop := event
			stop.Event = "unsubscribe"
			k.conn.WriteJSON(stop)
			return
		}
	}
}

func (k *Kraken) ReadHandler(wg *sync.WaitGroup, done context.Context, c chan domain.Candle) {
	defer func() {
		wg.Done()
		k.conn.Close()
	}()
	k.conn.SetReadDeadline(time.Now().Add(pongWait))
	k.conn.SetPongHandler(func(string) error { k.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	var version domain.Event
	k.conn.ReadJSON(&version)
	log.Println(version)

	candle := &domain.Candle{}
	tmp := domain.CandleSubscribe{}
	k.conn.ReadJSON(&tmp)
	candle.BuildCandle(tmp)

	for {
		select {
		case <-done.Done():
			return
		default:
			err := k.conn.ReadJSON(&tmp)
			if err != nil {
				break
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
