package exchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolver(t *testing.T) {
	c := MakeResolver()

	r := MakeResponse("test", "object", "id")
	rf := MakeResponseFuture("id")

	if assert.NotNil(t, c) {
		c.Track(rf)
		go c.Resolve(r)
		r2 := rf.Response()
		assert.Equal(t, r, r2)
	}

}
