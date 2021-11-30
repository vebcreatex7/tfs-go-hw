package domain

import "strings"

type Indicator struct {
	Fast   int    `json:"fast"`
	Slow   int    `json:"slow"`
	Signal int    `json:"signal"`
	Source string `json:"source"`
}

func (i *Indicator) IsValid() bool {
	if i.Fast < 0 || i.Slow < 0 || i.Signal < 0 {
		return false
	}
	i.Source = strings.ToUpper(i.Source)
	if i.Source != "O" && i.Source != "H" && i.Source != "L" && i.Source != "C" {
		return false
	}

	return true
}
