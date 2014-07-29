package inproc_test

import (
	"testing"
	"time"

	"github.com/qp/go"

	"github.com/qp/go/inproc"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {

	ps := inproc.NewPubSub()
	require.NotNil(t, ps)

	msgs := make(chan *qp.Message)
	data := []byte("testing")

	require.NoError(t, ps.Subscribe("channel", qp.HandlerFunc(func(msg *qp.Message) {
		msgs <- msg
	})))

	require.NoError(t, ps.Publish("channel", data))

	select {
	case msg := <-msgs:
		require.Equal(t, "channel", msg.Source)
		require.Equal(t, data, msg.Data)
	case <-time.After(100 * time.Millisecond):
		require.FailNow(t, "no message received")
	}

}
