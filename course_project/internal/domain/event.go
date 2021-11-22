package domain

type Event struct {
	Event      string        `json:"event"`
	Feed       string        `json:"feed,omitempty"`
	ProductIds []interface{} `json:"product_ids,omitempty"`
	Version    int           `json:"version,omitempty"`
}

func NewEvent(event string, period string, symbol string) Event {

	ans := Event{
		Event:      event,
		Feed:       "candle_trade_" + period,
		ProductIds: make([]interface{}, 1),
	}
	ans.ProductIds[0] = symbol
	return ans
}
