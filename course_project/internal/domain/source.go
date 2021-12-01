package domain

import "strings"

type Source struct {
	Source string `json:"source"`
}

func (i *Source) IsValid() bool {
	i.Source = strings.ToUpper(i.Source)
	if i.Source != "O" && i.Source != "H" && i.Source != "L" && i.Source != "C" {
		return false
	}
	return true
}
