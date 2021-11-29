package kraken

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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

var (
	ErrBadConnect = errors.New("kraken: bad subscribe")
)

type Kraken struct {
	conn               *websocket.Conn
	publicKey          string
	privateKey         string
	symbol             string
	period             domain.CandlePeriod
	amount             int
	openPositionAmount int
}

func NewKraken(public string, private string) *Kraken {
	return &Kraken{
		publicKey:  public,
		privateKey: private,
		amount:     1,
	}
}

/*
type KrakenService interface {
	SetSymbol(symbol string)
	GetSymbol() string
	SetPeriod(period domain.CandlePeriod)
	GetPeriod() domain.CandlePeriod
	WSConnect() error
	WSDisconnect() error
	CandlesFlow(*errgroup.Group, context.Context) <-chan domain.Candle
	GetOHLC(s string, p domain.CandlePeriod, n int64) ([]domain.Candle, error)
	GetOpenPositions() error
	SendOrderMkt(side string) (domain.Order, error)
}
*/
func (k *Kraken) SetSymbol(symbol string) {
	k.symbol = symbol
}

func (k *Kraken) GetSymbol() string {
	return k.symbol
}

func (k *Kraken) SetPeriod(period domain.CandlePeriod) {
	k.period = period
}

func (k *Kraken) GetPeriod() domain.CandlePeriod {
	return k.period
}

// Returns n previous candles with symbol s and period p
func (k *Kraken) GetOHLC(s string, p domain.CandlePeriod, n int64) ([]domain.Candle, error) {
	t := n * domain.GetPeriodInSec(p)
	from := time.Now().Unix() - t
	url := "https://demo-futures.kraken.com/api/charts/v1" + "/trade/" + s + "/" + string(p) + "?from=" + strconv.FormatInt(from, 10)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	d := domain.OHLC{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		return nil, err
	}

	if d.Error != "" {
		return nil, fmt.Errorf("GetOHLC error: %s", d.Error)
	}

	return d.Candles, nil
}

// Inits open positions
func (k *Kraken) GetOpenPositions() error {

	authent, err := k.genAuth("", "/api/v3/openpositions")
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, "https://demo-futures.kraken.com/derivatives/api/v3/openpositions", nil)
	if err != nil {
		return err
	}

	req.Header.Add("APIKey", k.publicKey)
	req.Header.Add("Authent", authent)

	c := http.Client{}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	d := domain.OpenPositions{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		return err
	}
	if d.Error != "" {
		return fmt.Errorf("GetOpenPositions error: %s", d.Error)
	}
	for i := range d.OpenPositions {
		if strings.EqualFold(d.OpenPositions[i].Symbol, k.symbol) {
			k.openPositionAmount = d.OpenPositions[i].Size
			if d.OpenPositions[i].Side == "short" {
				k.openPositionAmount *= -1
			}
			break
		}
	}
	return nil

}

// Sends mkt order which is executed at the market price
func (k *Kraken) SendOrderMkt(side string) (domain.Order, error) {
	postData := "orderType=mkt" + "&symbol=" + k.symbol + "&side=" + side + "&size=" + fmt.Sprintf("%d", k.amount)
	authent, err := k.genAuth(postData, "/api/v3/sendorder")
	if err != nil {
		return domain.Order{}, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://demo-futures.kraken.com/derivatives/api/v3/sendorder"+"?"+postData, nil)
	if err != nil {
		return domain.Order{}, err
	}

	req.Header.Add("APIKey", k.publicKey)
	req.Header.Add("Authent", authent)

	c := http.Client{}

	resp, err := c.Do(req)
	if err != nil {
		return domain.Order{}, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Order{}, err
	}

	d := domain.Order{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		return domain.Order{}, err
	}

	if d.Error != "" {
		return domain.Order{}, fmt.Errorf("SendOrderMkt error: %s", d.Error)
	}

	return d, nil

}

// Connect via websocket for the candles flow
func (k *Kraken) WSConnect() error {
	k.conn = nil
	urlWs := "wss://demo-futures.kraken.com/ws/v1?chart"
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

// Disconnects
func (k *Kraken) WSDisconnect() error {
	err := k.conn.Close()
	k.conn = nil
	if err != nil {
		return err
	}

	return nil
}

// First part of pipline, returns candles channel
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
			return err
		}

		// Ping
		for {
			select {
			case <-ping.C:
				err = k.conn.WriteMessage(websocket.PingMessage, nil)
				if err != nil {
					return err
				}
			case <-ctx.Done():
				stop := event
				stop.Event = "unsubscribe"
				err = k.conn.WriteJSON(stop)
				if err != nil {
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
			return err
		}

		// Subscribed
		var event domain.Event
		err = k.conn.ReadJSON(&event)
		if err != nil {
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

		// There is a lot of problems for example the repeatedly sends unique candles.
		// To fight with this, this variable has been introduced.
		var gotVithZeroVolume bool = false

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
				if tmp.C.Time < candle.Time {
					continue
				} else if tmp.C.Time == candle.Time {
					if gotVithZeroVolume && tmp.C.Volume == 0 {
						continue
					}
					candle.Close = tmp.C.Close
					if tmp.C.Low < candle.Low {
						candle.Low = tmp.C.Low
					}
					if tmp.C.High > candle.High {
						candle.High = tmp.C.High
					}
					candle.Volume = tmp.C.Volume
					if tmp.C.Volume == 0. {
						gotVithZeroVolume = true
					}
				} else {
					c <- *candle
					candle.BuildCandle(tmp)
					gotVithZeroVolume = false
				}
			}
		}

	})
	return c
}

// Used for private REST API
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
