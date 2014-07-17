package event

import (
	"strings"
	"sync"
	"time"

	"github.com/qp/go/transports"
)

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
	callback transports.MessageFunc
	wrapped  transports.EventTransport
}

var queue = make(chan *transports.BinaryMessage)
var channels = map[string][]*InProc{}
var lock sync.RWMutex

func processMessages() {
	go func() {
		for {
			select {
			case bm := <-queue:
				lock.RLock()
				for channel, procs := range channels {
					if (strings.HasSuffix(channel, "*") &&
						strings.HasPrefix(channel, bm.Channel)) ||
						channel == bm.Channel {
						for _, instance := range procs {
							go instance.callback(bm)
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
func MakeInProc(wrapped transports.EventTransport) transports.EventTransport {
	return &InProc{wrapped: wrapped}
}

// ListenFor instructs InProc to listen for a message for the given channel
func (i *InProc) ListenFor(channel string) {
	// listen on a channel
	lock.Lock()
	channels[channel] = append(channels[channel], i)
	lock.Unlock()
	if i.wrapped != nil {
		i.wrapped.ListenFor(channel)
	}
}

// ListenForChildren instructs InProc to listen for a message for the given channel
// and all its children
func (i *InProc) ListenForChildren(channel string) {
	channel += "*"
	// listen on a channel
	lock.Lock()
	channels[channel] = append(channels[channel], i)
	lock.Unlock()
	if i.wrapped != nil {
		i.wrapped.ListenFor(channel)
	}
}

// OnMessage assigns a callback function to be called when a message
// is received on this transport
func (i *InProc) OnMessage(messageFunc transports.MessageFunc) {
	// assign the callback to be called
	i.callback = messageFunc
	if i.wrapped != nil {
		i.wrapped.OnMessage(messageFunc)
	}
}

// Send sends a message into the transport
func (i *InProc) Send(channel string, message []byte) error {
	lock.RLock()
	_, ok := channels[channel]
	lock.RUnlock()
	if ok {
		queue <- &transports.BinaryMessage{Channel: channel, Data: message}
	} else {
		if i.wrapped != nil {
			return i.wrapped.Send(channel, message)
		}
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

// SetTimeout is a no-op for the InProc transport
func (i *InProc) SetTimeout(timeout time.Duration) {
}
