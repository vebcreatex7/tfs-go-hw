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
	Pair []interface{} `json:"symbol"`
}

func (s *Symbol) GetValid() []string {
	valid := make([]string, 0)
	for i := range s.Pair {
		val, ok := s.Pair[i].(string)
		if ok {
			isValid := false
			for j := range AvailableSymbols {
				if strings.ToLower(val) == AvailableSymbols[j] {
					isValid = true
					break
				}
			}
			if isValid {
				valid = append(valid, val)
			}
		}
	}
	return valid
}
