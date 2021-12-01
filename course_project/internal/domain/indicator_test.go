package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidIndicator_Testify(t *testing.T) {
	i := &Indicator{
		Fast:   12,
		Slow:   26,
		Signal: 9,
		Source: "C",
	}

	assert.Equal(t, true, i.IsValid())
}
