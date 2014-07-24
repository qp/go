package redis_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/redis"
	"github.com/stretchr/testify/assert"
)

func TestRequests(t *testing.T) {

	if !initRedis() {
		t.Skip("Cannot start redis - skipping.")
	}

	r := redis.NewReqTransport("127.0.0.1:6379")
	a := assert.New(t)
	channel := "test"
	mc := make(chan *qp.Message)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	r.ListenFor(channel)
	r.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})
	r.Start()
	r.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	r.Stop()

	// make sure we can start again after stopping
	r.ListenFor(channel)
	r.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})

	r.Start()
	r.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	r.Stop()

	stopRedis()

}
