package inproc_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/inproc"
	"github.com/stretchr/testify/assert"
)

func TestNewEvents(t *testing.T) {

	ip := inproc.NewPubSubTransport(nil)
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

}

func TestInProcMultiple(t *testing.T) {

	ip := inproc.NewPubSubTransport(nil)

	a := assert.New(t)

	channel := "test.event"
	mc := make(chan *qp.Message)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	ip.ListenFor(channel)
	ip.ListenForChildren(channel)
	ip.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})

	ip.Start()

	ip2 := inproc.NewPubSubTransport(nil)
	ip2.ListenForChildren(channel)
	ip2.OnMessage(func(bm *qp.Message) {
		mc <- bm
	})

	ip2.Start()

	ip.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Source)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip2.Stop()
}
