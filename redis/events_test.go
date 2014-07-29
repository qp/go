package redis_test

import (
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/redis"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {

	ensureRedis(t)

	r := redis.NewPubSubTransport("127.0.0.1:6379")
	data := []byte(`testing`)
	hit := make(chan struct{})

	r.ListenFor("test.event")
	r.ListenForChildren("test.event")
	r.OnMessage(func(bm *qp.Message) {
		assert.Equal(t, "test.event", bm.Source)
		assert.Equal(t, data, bm.Data)
		hit <- struct{}{}
	})

	r.Start()

	time.Sleep(10 * time.Millisecond)

	publisher := redis.NewPubSubTransport("127.0.0.1:6379")
	publisher.Start()
	publisher.Send("test.event", data)
	publisher.Stop()

	count := 0
loop:
	for {
		select {
		case <-hit:
			count++
			if count == 2 {
				break loop
			}
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "Timed out while waiting for message")
			break loop
		}
	}

	assert.Equal(t, 2, count)

	r.Stop()

}
