package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/qp/go/transports"
	"github.com/stretchr/testify/assert"
)

// ensure InProc conforms to Transport interface
var _ transports.Transport = (*InProc)(nil)

func TestInProc(t *testing.T) {

	ip := MakeInProc(nil)
	a := assert.New(t)
	channel := "test"
	mc := make(chan *transports.BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	ip.ListenFor(channel)
	ip.OnMessage(func(bm *transports.BinaryMessage) {
		mc <- bm
	})
	ip.Start()
	ip.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Channel)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip.Stop()

}

func TestInProcMultiple(t *testing.T) {

	ip := MakeInProc(nil)

	a := assert.New(t)

	channel := "test.event"
	channel2 := "test.event.*"
	mc := make(chan *transports.BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	ip.ListenFor(channel)
	ip.OnMessage(func(bm *transports.BinaryMessage) {
		mc <- bm
	})

	ip.Start()
	ip.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Channel)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip2 := MakeInProc(nil)
	ip2.ListenFor(channel2)
	ip2.OnMessage(func(bm *transports.BinaryMessage) {
		mc <- bm
	})

	ip2.Start()

	ip.Send(channel, data)

	// ip.Stop() should have no effect on ip2
	ip.Stop()

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Channel)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip2.Stop()
}
