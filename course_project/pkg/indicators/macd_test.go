package indicators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tfs-go-hw/course_project/internal/domain"
)

func TestPredict_Testify(t *testing.T) {
	m := &Macd{
		macd: domain.Candle{
			Open:  "1.",
			High:  "1.",
			Low:   "1.",
			Close: "1.",
		},
		macdPrev: domain.Candle{
			Open:  "-1.",
			High:  "-1.",
			Low:   "-1.",
			Close: "-1.",
		},
		signal: domain.Candle{
			Open:  "-1.",
			High:  "-1.",
			Low:   "-1.",
			Close: "-1.",
		},
		signalPrev: domain.Candle{
			Open:  "1.",
			High:  "1.",
			Low:   "1.",
			Close: "1.",
		},
	}

	expect := domain.Buy

	m.source = 'O'
	assert.Equal(t, expect, m.predict())
	m.source = 'H'
	assert.Equal(t, expect, m.predict())
	m.source = 'L'
	assert.Equal(t, expect, m.predict())
	m.source = 'C'
	assert.Equal(t, expect, m.predict())

}
