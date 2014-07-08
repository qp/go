package transports

import "sync"

type instanceID uint64

// InProc is the InProc implementation of the
// Transport interface.
//
// The InProc transport implements everything using in-process
// communication methods. This is useful for creating a
// service style system inside a single application for
// initial development and testing. When ready, it requires
// minimal effort to split each service into separate
// processes.
//
// InProc should only be used for request and reply, not
// events.
type InProc struct {
	callback MessageFunc
	wrapped  Transport
}

var queue = make(chan *BinaryMessage)
var channels = map[string][]*InProc{}
var lock sync.RWMutex

func processMessages() {
	go func() {
		for {
			select {
			case bm := <-queue:
				lock.RLock()
				if instances, ok := channels[bm.Channel]; ok {
					for _, instance := range instances {
						go instance.callback(bm)
					}
				} else {
					// nothing found. forward to wrapped transport
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
func MakeInProc(wrapped Transport) Transport {
	return &InProc{wrapped: wrapped}
}

// ListenFor instructs InProc to deliver a message for the given channel
func (i *InProc) ListenFor(channel string) {
	// listen on a channel
	lock.Lock()
	channels[channel] = append(channels[channel], i)
	lock.Unlock()
	i.wrapped.ListenFor(channel)
}

// OnMessage assigns a callback function to be called when a message
// is received on this transport
func (i *InProc) OnMessage(messageFunc MessageFunc) {
	// assign the callback to be called
	i.callback = messageFunc
	i.wrapped.OnMessage(messageFunc)
}

// Send sends a message into the transport
func (i *InProc) Send(channel string, message []byte) error {
	lock.RLock()
	_, ok := channels[channel]
	lock.RUnlock()
	if ok {
		queue <- &BinaryMessage{Channel: channel, Data: message}
	} else {
		return i.wrapped.Send(channel, message)
	}
	return nil
}

// Start is a no-op for the InProc transport.
func (i *InProc) Start() error {
	return nil
}

// Stop is a no-op for the InProc transport.
func (i *InProc) Stop() {
}
