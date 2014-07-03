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

	ip := &InProc{}
	a := assert.New(t)
	topic := "test"
	mc := make(chan *BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	err := ip.ListenFor(topic, func(bm *BinaryMessage) {
		mc <- bm
	})
	ip.Start()
	ip.Send(topic, data)
	if a.NoError(err) {
		select {
		case bm := <-mc:
			a.Equal(topic, bm.Topic)
			a.Equal(data, bm.Data)
		case <-time.After(100 * time.Millisecond):
			a.Fail("No message received!")
		}
	}
	ip.Stop()

	// make sure we can start again after stopping
	err = ip.ListenFor(topic, func(bm *BinaryMessage) {
		mc <- bm
	})
	ip.Start()
	ip.Send(topic, data)
	if a.NoError(err) {
		select {
		case bm := <-mc:
			a.Equal(topic, bm.Topic)
			a.Equal(data, bm.Data)
		case <-time.After(100 * time.Millisecond):
			a.Fail("No message received!")
		}
	}
	ip.Stop()

}

func TestInProcMultiple(t *testing.T) {

	ip := NewInProc()

	a := assert.New(t)

	topic := "test"
	mc := make(chan *BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	err := ip.ListenFor(topic, func(bm *BinaryMessage) {
		mc <- bm
	})
	ip.Start()
	ip.Send(topic, data)
	if a.NoError(err) {
		select {
		case bm := <-mc:
			a.Equal(topic, bm.Topic)
			a.Equal(data, bm.Data)
		case <-time.After(100 * time.Millisecond):
			a.Fail("No message received!")
		}
	}

	ip2 := NewInProc()
	err = ip2.ListenFor(topic, func(bm *BinaryMessage) {
		mc <- bm
	})
	ip2.Start()

	// ip.Stop() should have no effect on ip2
	ip.Stop()

	ip2.Send(topic, data)
	if a.NoError(err) {
		select {
		case bm := <-mc:
			a.Equal(topic, bm.Topic)
			a.Equal(data, bm.Data)
		case <-time.After(100 * time.Millisecond):
			a.Fail("No message received!")
		}
	}
	ip2.Stop()
}
