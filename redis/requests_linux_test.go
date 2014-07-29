package redis_test

import (
	"encoding/json"
	"fmt"
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
	r.SetTimeout(1 * time.Millisecond)
	r.Start()
	r.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	select {
	case <-r.Stop():
		fmt.Println("received stop 1")
	case <-time.After(2 * time.Second):
		a.Fail("Stop timed out")
	}

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

	select {
	case <-r.Stop():
		fmt.Println("received stop 2")
	case <-time.After(2 * time.Second):
		a.Fail("Stop timed out")
	}

}

func TestRequestsTimeout(t *testing.T) {

	if !initRedis() {
		t.Skip("Cannot start redis - skipping.")
	}

	r := redis.NewReqTransport("127.0.0.1:6379")
	r.SetTimeout(1 * time.Millisecond)
	a := assert.New(t)
	channel := "timeout"
	mc := make(chan *qp.Message)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	r.ListenFor(channel)
	r.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})
	fmt.Println("starting timeout r")
	r.Start()

	// send the message in
	go func() {
		time.Sleep(1100 * time.Millisecond)
		r.Send(channel, data)
	}()

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(1500 * time.Millisecond): // timeout
		a.Fail("No message received!")
	}

	<-r.Stop()

}
