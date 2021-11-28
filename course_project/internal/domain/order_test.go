package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRecordOrder_Testify(t *testing.T) {
	ope := orderPriorExecution{
		Symbol: "pi_xbtusd",
		Side:   "buy",
	}

	oe := orderEvent{
		Price:               5555.5,
		Amount:              1,
		OrderPriorExecution: ope,
	}

	oes := make([]orderEvent, 1)
	oes[0] = oe

	ss := sendStatus{
		ReceivedTime: "2021-11-28T11:45:26.371Z",
		OrderEvents:  oes,
	}

	o := Order{
		SendStatus: ss,
	}

	layout := "2006-01-02T15:04:05.000Z"
	str := o.SendStatus.ReceivedTime
	ts, _ := time.Parse(layout, str)

	loc, _ := time.LoadLocation("Europe/Moscow")

	ts = ts.In(loc)

	expect := &RecordOrder{
		TS:     ts,
		Symbol: "pi_xbtusd",
		Side:   "buy",
		Size:   1,
		Price:  5555.5,
	}

	got, err := NewRecordOrder(&o)
	if err != nil {
		assert.Error(t, err, "got error")
	}

	assert.Equal(t, expect, got)

}
