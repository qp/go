package qp

import (
	"testing"
	"time"

	"github.com/qp/go/messages"
	"github.com/stretchr/testify/assert"
)

func TestQPResponse(t *testing.T) {

	r := newResponse(10 * time.Millisecond)

	go func() {
		r.response <- messages.NewMessage("test", "data", "service")
	}()

	msg := r.Message()

	if assert.False(t, msg.HasError()) {
		assert.Equal(t, msg.Data, "data")
	}

}

func TestQPResponseTimeout(t *testing.T) {

	r := newResponse(10 * time.Millisecond)

	msg := r.Message()

	assert.True(t, msg.HasError())

}
