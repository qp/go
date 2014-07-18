package exchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventMapper(t *testing.T) {

	e := MakeEventMapper()

	e.Track("test.event.one", nil)
	e.Track("test.event*", nil)

	i := e.Find("test.event.one")
	assert.Equal(t, 2, len(i))

}
