package transports

import (
	"sync"

	"code.google.com/p/go-uuid/uuid"
)

type instanceID string

// InProc is the InProc implementation of the
// Transport interface.
//
// The InProc transport implements everything using in-process
// communication methods. This is useful for creating a
// service style system inside a single application for
// initial development and testing. When ready, it requires
// minimal effort to split each service into separate
// processes.
type InProc struct {
	id instanceID
}

type inProcMapper struct {
	channels []string
	callback MessageFunc
}

var queue = make(chan *BinaryMessage)
var maps = map[instanceID]*inProcMapper{}
var lock sync.RWMutex

func processMessages() {
	go func() {
		for {
			select {
			case bm := <-queue:
				lock.RLock()
				for _, mapper := range maps {
					for _, channel := range mapper.channels {
						if channel == bm.Channel {
							go mapper.callback(bm)
						}
					}
				}
				lock.RUnlock()
			}
		}
	}()
}

func init() {
	processMessages()
}

// MakeInProc creates a new instance of InProc
func MakeInProc() Transport {
	return &InProc{id: instanceID(uuid.New())}
}

// ListenFor instructs InProc to deliver a message for the given channel
func (i *InProc) ListenFor(channel string) {
	// listen on a channel
	if _, ok := maps[i.id]; !ok {
		maps[i.id] = &inProcMapper{}
	}
	lock.Lock()
	maps[i.id].channels = append(maps[i.id].channels, channel)
	lock.Unlock()
}

// OnMessage assigns a callback function to be called when a message
// is received on this transport
func (i *InProc) OnMessage(messageFunc MessageFunc) {
	// assign the callback to be called
	if _, ok := maps[i.id]; !ok {
		maps[i.id] = &inProcMapper{}
	}
	lock.Lock()
	maps[i.id].callback = messageFunc
	lock.Unlock()
}

// Send sends a message into the transport
func (i *InProc) Send(channel string, message []byte) error {
	queue <- &BinaryMessage{Channel: channel, Data: message}
	return nil
}

// Start is a no-op for the InProc transport.
func (i *InProc) Start() error {
	return nil
}

// Stop removes all callbacks for this instance of InProc
func (i *InProc) Stop() {
}
