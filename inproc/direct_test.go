package inproc_test

import (
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/inproc"
	"github.com/stretchr/pat/stop"
	"github.com/stretchr/testify/require"
)

func TestDirect(t *testing.T) {

	d := inproc.NewDirect()
	d.Start()
	defer func() {
		d.Stop(stop.NoWait)
		<-d.StopChan()
	}()
	require.NotNil(t, d)

	msgs := make(chan *qp.Message)
	data := []byte("testing")

	require.NoError(t, d.OnMessage("channel", qp.HandlerFunc(func(msg *qp.Message) {
		msgs <- msg
	})))

	require.NoError(t, d.Send("channel", data))

	select {
	case msg := <-msgs:
		require.Equal(t, "channel", msg.Source)
		require.Equal(t, data, msg.Data)
	case <-time.After(100 * time.Millisecond):
		require.FailNow(t, "no message received")
	}

}
