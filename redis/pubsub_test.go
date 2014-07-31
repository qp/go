package redis_test

import (
	"testing"
	"time"

	"github.com/qp/go"

	"github.com/qp/go/redis"
	"github.com/stretchr/pat/stop"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {

	ensureRedis(t)

	publisher := redis.NewPubSub("127.0.0.1:6379")
	ps := redis.NewPubSub("127.0.0.1:6379")
	ps2 := redis.NewPubSub("127.0.0.1:6379")
	require.NotNil(t, ps)
	require.NotNil(t, ps2)

	defer func() {
		publisher.Stop(stop.NoWait)
		ps.Stop(stop.NoWait)
		ps2.Stop(stop.NoWait)
		<-publisher.StopChan()
		<-ps.StopChan()
		<-ps2.StopChan()
	}()

	msgs := make(chan *qp.Message)
	data := []byte("testing")

	require.NoError(t, ps.Subscribe("channel", qp.HandlerFunc(func(msg *qp.Message) {
		msgs <- msg
	})))
	require.NoError(t, ps2.Subscribe("channel", qp.HandlerFunc(func(msg *qp.Message) {
		msgs <- msg
	})))

	require.NoError(t, publisher.Start())
	require.NoError(t, ps.Start())
	require.NoError(t, ps2.Start())

	time.Sleep(100 * time.Millisecond)

	require.NoError(t, publisher.Publish("channel", data))

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
			case <-time.After(1000 * time.Millisecond):
				require.FailNow(t, "no message received")
				return
			}
		}
	}()

}
