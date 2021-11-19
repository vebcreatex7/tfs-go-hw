package domain

type internalCandle struct {
	Time   int     `json:"time"`
	Open   string  `json:"open"`
	High   string  `json:"high"`
	Low    string  `json:"low"`
	Close  string  `json:"close"`
	Volume float64 `json:"volume"`
}

type Candle struct {
	Feed      string         `json:"feed"`
	Candle    internalCandle `json:"candle,omitempty"`
	ProductId string         `json:"product_id"`
}
