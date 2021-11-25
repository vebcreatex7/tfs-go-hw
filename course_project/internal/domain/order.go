package domain

type Action string

var (
	Hold Action = "hold"
	Buy  Action = "buy"
	Sell Action = "sell"
)

type orderPriorExecution struct {
	OrderId    string  `json:"orderId,omitempty"`
	Type       string  `json:"type,omitempty"`
	Symbol     string  `json:"symbol,omitempty"`
	Side       string  `json:"side,omitempty"`
	Quantity   int     `json:"quantity,omitempty"`
	LimitPrice float64 `json:"limitPrice,omitempty"`
	Timestamp  string  `json:"timestamp,omitempty"`
}

type orderEvent struct {
	ExecutionId         string              `json:"executionId,omitempty"`
	Price               float64             `json:"price,omitempty"`
	Amount              int                 `json:"amount,omitempty"`
	OrderPriorExecution orderPriorExecution `json:"orderPriorExecution,omitempty"`
	Type                string              `json:"type,omitempty"`
}

type sendStatus struct {
	OrderId      string       `json:"order_id,omitempty"`
	СliOrdId     string       `json:"cliOrdId,omitempty"`
	Status       string       `json:"status,omitempty"`
	ReceivedTime string       `json:"receivedTime,omitempty"`
	OrderEvents  []orderEvent `json:"orderEvents,omitempty"`
}

type Order struct {
	Result     string     `json:"result,omitempty"`
	SendStatus sendStatus `json:"sendStatus,omitempty"`
	ServerTime string     `json:"serverTime,omitempty"`
	Error      string     `json:"error,omitempty"`
}
