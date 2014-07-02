package transport

import (
	"sync"

	"code.google.com/p/go-uuid/uuid"
)

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
	id string
}

var callbacks = map[string]map[string][]MessageFunc{}
var queue = make(chan *BinaryMessage)
var lock sync.RWMutex

func processMessages() {
	go func() {
		for {
			select {
			case bm := <-queue:
				lock.RLock()
				for _, instance := range callbacks {
					for _, cb := range instance[bm.topic] {
						cb(bm)
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

// NewInProc creates a new instance of InProc
func NewInProc() *InProc {
	return &InProc{id: uuid.New()}
}

// ListenFor instructs InProc to deliver a message for the given topic
func (r *InProc) ListenFor(topic string, callback MessageFunc) error {
	lock.Lock()
	if _, ok := callbacks[r.id]; !ok {
		callbacks[r.id] = map[string][]MessageFunc{}
	}
	callbacks[r.id][topic] = append(callbacks[r.id][topic], callback)
	lock.Unlock()
	return nil
}

// Send sends a message into the transport
func (r *InProc) Send(message *BinaryMessage) error {
	queue <- message
	return nil
}

// Start is a no-op for the InProc transport.
func (r *InProc) Start() {
}

// Stop removes all callbacks for this instance of InProc
func (r *InProc) Stop() {
	lock.Lock()
	delete(callbacks, r.id)
	lock.Unlock()
}
