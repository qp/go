package transports

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ensure InProc conforms to Transport interface
var _ Transport = (*InProc)(nil)

func TestInProc(t *testing.T) {

	ip := MakeInProc(MakeLog(true))
	a := assert.New(t)
	channel := "test"
	mc := make(chan *BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	ip.ListenFor(channel)
	ip.OnMessage(func(bm *BinaryMessage) {
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

	// make sure we can start again after stopping
	ip.ListenFor(channel)
	ip.OnMessage(func(bm *BinaryMessage) {
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

	ip := MakeInProc(MakeLog(true))

	a := assert.New(t)

	channel := "test"
	mc := make(chan *BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	ip.ListenFor(channel)
	ip.OnMessage(func(bm *BinaryMessage) {
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

	ip2 := MakeInProc(MakeLog(true))
	ip2.ListenFor(channel)
	ip2.OnMessage(func(bm *BinaryMessage) {
		mc <- bm
	})

	ip2.Start()

	// ip.Stop() should have no effect on ip2
	ip.Stop()

	ip2.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Channel)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	ip2.Stop()
}
