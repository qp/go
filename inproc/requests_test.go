package inproc_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/inproc"
	"github.com/stretchr/testify/assert"
)

func TestNewRequests(t *testing.T) {

	ip := inproc.NewReqTransport(nil)
	a := assert.New(t)
	channel := "test"
	mc := make(chan *qp.Message)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	ip.ListenFor(channel)
	ip.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})
	ip.Start()
	ip.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip.Stop()

	// make sure we can start again after stopping
	ip.ListenFor(channel)
	ip.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})

	ip.Start()
	ip.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip.Stop()

}

func TestNewRequestsMultiple(t *testing.T) {

	ip := inproc.NewReqTransport(nil)

	a := assert.New(t)

	channel := "test"
	mc := make(chan *qp.Message)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	ip.ListenFor(channel)
	ip.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})

	ip.Start()
	ip.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip2 := inproc.NewReqTransport(nil)
	ip2.ListenFor(channel)
	ip2.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})

	ip2.Start()

	// ip.Stop() should have no effect on ip2
	ip.Stop()

	ip2.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip2.Stop()
}
