package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsVaid_Testify(t *testing.T) {
	p := &Period{
		Period: CandlePeriod12h,
	}

	assert.Equal(t, true, p.IsValid())
}

func TestGetPeriodInSec_Testify(t *testing.T) {
	p := CandlePeriod12h
	var expect int64 = 43200
	assert.Equal(t, expect, GetPeriodInSec(p))

}
