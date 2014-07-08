package qp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolver(t *testing.T) {
	c := makeResolver()

	r := makeResponse("test", "object", "id")
	rf := makeResponseFuture("id")

	if assert.NotNil(t, c) {
		c.track(rf)
		go c.resolve(r)
		r2 := rf.Response()
		assert.Equal(t, r, r2)
	}

}
