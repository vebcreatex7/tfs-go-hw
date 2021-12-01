package domain

import "strings"

var AvailableSymbols [5]string = [5]string{
	"pi_xbtusd",
	"pi_ethusd",
	"pi_ltcusd",
	"pi_xrpusd",
	"pi_bchusd",
}

type Symbol struct {
	Symbol string `json:"symbol"`
}

func (s *Symbol) IsValid() bool {
	isValid := false
	for i := range AvailableSymbols {
		if strings.ToLower(s.Symbol) == AvailableSymbols[i] {
			isValid = true
			break
		}
	}
	return isValid

}
