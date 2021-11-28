package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid_Testify(t *testing.T) {
	s := &Symbol{
		Symbol: "pi_xbtusd",
	}

	assert.Equal(t, true, s.IsValid())

}
