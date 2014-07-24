package qp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnique(t *testing.T) {
	u = 0
	assert.Equal(t, unique(), 1)
	assert.Equal(t, unique(), 2)
	assert.Equal(t, unique(), 3)
	// assert.Equal(t, uniqueStr(), "4")
	// assert.Equal(t, uniqueStr(), "5")
	// assert.Equal(t, uniqueStr(), "6")
}
