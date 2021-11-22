package domain

type Candle struct {
	Time   int     `json:"time"`
	Open   string  `json:"open"`
	High   string  `json:"high"`
	Low    string  `json:"low"`
	Close  string  `json:"close"`
	Volume float64 `json:"volume"`
}

func (candle *Candle) BuildCandle(tmp CandleSubscribe) {
	candle.Time = tmp.C.Time
	candle.Open = tmp.C.Open
	candle.High = tmp.C.High
	candle.Low = tmp.C.Low
	candle.Close = tmp.C.Close
	candle.Volume = tmp.C.Volume
}

type CandleSubscribe struct {
	Feed      string `json:"feed"`
	C         Candle `json:"candle,omitempty"`
	ProductId string `json:"product_id"`
}
