package inproc_test

import (
	"testing"
	"time"

	"github.com/qp/go"

	"github.com/qp/go/inproc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {

	ps := inproc.NewPubSub()
	ps2 := inproc.NewPubSub()
	require.NotNil(t, ps)
	require.NotNil(t, ps2)

	msgs := make(chan *qp.Message)
	data := []byte("testing")

	require.NoError(t, ps.Subscribe("channel", qp.HandlerFunc(func(msg *qp.Message) {
		msgs <- msg
	})))
	require.NoError(t, ps2.Subscribe("channel", qp.HandlerFunc(func(msg *qp.Message) {
		msgs <- msg
	})))

	require.NoError(t, ps.Publish("channel", data))

	count := 0
	func() {
		for {
			select {
			case msg := <-msgs:
				require.Equal(t, "channel", msg.Source)
				require.Equal(t, data, msg.Data)
				count++
				if count == 2 {
					return
				}
			case <-time.After(100 * time.Millisecond):
				require.FailNow(t, "no message received")
				return
			}
		}
	}()

	assert.Equal(t, 2, count)

}
