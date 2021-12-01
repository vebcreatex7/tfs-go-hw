package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRecordOrder_Testify(t *testing.T) {
	ope := OrderPriorExecution{
		Symbol: "pi_xbtusd",
		Side:   "buy",
	}

	oe := OrderEvent{
		Price:               5555.5,
		Amount:              1,
		OrderPriorExecution: ope,
	}

	oes := make([]OrderEvent, 1)
	oes[0] = oe

	ss := SendStatus{
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
