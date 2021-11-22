package domain

type CandlePeriod string

const (
	CandlePeriod1m  CandlePeriod = "1m"
	CandlePeriod5m  CandlePeriod = "5m"
	CandlePeriod15m CandlePeriod = "15m"
	CandlePeriod30m CandlePeriod = "30m"
	CandlePeriod1h  CandlePeriod = "1h"
	CandlePeriod4h  CandlePeriod = "4h"
	CandlePeriod12h CandlePeriod = "12h"
	CandlePeriod1d  CandlePeriod = "1d"
	CandlePeriod1w  CandlePeriod = "1w"
)

type Period struct {
	Period CandlePeriod `json:"period"`
}

func (p *Period) IsValid() bool {
	isValid := true
	switch p.Period {
	case "1m":
		p.Period = CandlePeriod1m
	case "5m":
		p.Period = CandlePeriod5m
	case "15m":
		p.Period = CandlePeriod15m
	case "30m":
		p.Period = CandlePeriod30m
	case "1h":
		p.Period = CandlePeriod1h
	case "4h":
		p.Period = CandlePeriod4h
	case "12h":
		p.Period = CandlePeriod12h
	case "1d":
		p.Period = CandlePeriod1d
	case "1w":
		p.Period = CandlePeriod1w
	default:
		isValid = false
	}
	return isValid
}
