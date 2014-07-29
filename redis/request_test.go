package redis_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/redis"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {

	ensureRedis(t)
	url := "127.0.0.1:6379"
	channel := "testrequest"
	recvdMsgs := make(chan *qp.Message)
	testData, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	r := redis.NewReqTransport(url)
	assert.NotNil(t, r)
	r.ListenFor(channel)
	r.OnMessage(func(bm *qp.Message) {
		recvdMsgs <- bm
	})
	assert.NoError(t, r.Start())
	r.Send(channel, testData)

	select {
	case bm := <-recvdMsgs:
		assert.Equal(t, channel, bm.Source)
		assert.Equal(t, testData, bm.Data)
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "(timeout) No message received!")
	}

	select {
	case <-r.Stop():
	case <-time.After(2 * time.Second):
		assert.Fail(t, "(timeout) Didn't Stop")
	}

}
