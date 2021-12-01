package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent_Testify(t *testing.T) {
	e := Event{
		Event:      "sucsess",
		Feed:       "candles_trade_1m",
		ProductIds: []interface{}{"pi_xbtusd"},
	}

	assert.Equal(t, e, NewEvent("sucsess", "1m", "pi_xbtusd"))
}
