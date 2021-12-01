package domain

import "time"

type Action string

var (
	Hold Action = "hold"
	Buy  Action = "buy"
	Sell Action = "sell"
)

type OrderPriorExecution struct {
	OrderId    string  `json:"orderId,omitempty"`
	Type       string  `json:"type,omitempty"`
	Symbol     string  `json:"symbol,omitempty"`
	Side       string  `json:"side,omitempty"`
	Quantity   int     `json:"quantity,omitempty"`
	LimitPrice float64 `json:"limitPrice,omitempty"`
	Timestamp  string  `json:"timestamp,omitempty"`
}

type OrderEvent struct {
	ExecutionId         string              `json:"executionId,omitempty"`
	Price               float64             `json:"price,omitempty"`
	Amount              int                 `json:"amount,omitempty"`
	OrderPriorExecution OrderPriorExecution `json:"orderPriorExecution,omitempty"`
	Type                string              `json:"type,omitempty"`
}

type SendStatus struct {
	OrderId      string       `json:"order_id,omitempty"`
	СliOrdId     string       `json:"cliOrdId,omitempty"`
	Status       string       `json:"status,omitempty"`
	ReceivedTime string       `json:"receivedTime,omitempty"`
	OrderEvents  []OrderEvent `json:"orderEvents,omitempty"`
}

type Order struct {
	Result     string     `json:"result,omitempty"`
	SendStatus SendStatus `json:"sendStatus,omitempty"`
	ServerTime string     `json:"serverTime,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type RecordOrder struct {
	TS     time.Time
	Symbol string
	Side   string
	Size   int
	Price  float64
}

func NewRecordOrder(o *Order) (*RecordOrder, error) {
	layout := "2006-01-02T15:04:05.000Z"
	str := o.SendStatus.ReceivedTime
	t, err := time.Parse(layout, str)

	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, err
	}
	t = t.In(loc)

	symb := o.SendStatus.OrderEvents[0].OrderPriorExecution.Symbol
	side := o.SendStatus.OrderEvents[0].OrderPriorExecution.Side
	size := o.SendStatus.OrderEvents[0].Amount
	price := o.SendStatus.OrderEvents[0].Price

	return &RecordOrder{
		TS:     t,
		Symbol: symb,
		Side:   side,
		Size:   size,
		Price:  price,
	}, nil
}
