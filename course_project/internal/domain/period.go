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

func GetPeriodInSec(p CandlePeriod) int64 {
	switch p {
	case CandlePeriod1m:
		return 60
	case CandlePeriod5m:
		return 5 * 60
	case CandlePeriod15m:
		return 15 * 60
	case CandlePeriod30m:
		return 30 * 60
	case CandlePeriod1h:
		return 1 * 60 * 60
	case CandlePeriod4h:
		return 4 * 60 * 60
	case CandlePeriod12h:
		return 12 * 60 * 60
	case CandlePeriod1d:
		return 1 * 24 * 60 * 60
	case CandlePeriod1w:
		return 7 * 24 * 60 * 60
	}
	return 0
}
