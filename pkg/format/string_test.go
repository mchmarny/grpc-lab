package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringReversal(t *testing.T) {
	s := "test"

	t.Run("string reversal", func(t *testing.T) {
		v := ReverseString(s)
		assert.Equal(t, "tset", v)
	})
}
