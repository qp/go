package inproc

import (
	"sync"
	"time"

	"github.com/stretchr/pat/stop"

	"github.com/qp/go"
)

// requests is the InProc implementation of the
// qp.RequestTransport interface.
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
type requests struct {
	callback qp.MessageFunc
	wrapped  qp.RequestTransport
}

var reqQueue = make(chan *qp.Message)
var reqChannels = map[string][]*requests{}
var reqLock sync.RWMutex

func processRequestsMessages() {
	go func() {
		for {
			select {
			case bm := <-reqQueue:
				reqLock.RLock()
				for _, instance := range reqChannels[bm.Source] {
					go instance.callback(bm)
				}
				reqLock.RUnlock()
			}
		}
	}()
}

func init() {
	processRequestsMessages()
}

// NewReqTransport creates a new in-process qp.RequestTransport.
func NewReqTransport(wrapped qp.RequestTransport) qp.RequestTransport {
	return &requests{wrapped: wrapped}
}

// ListenFor instructs InProc to deliver a message for the given channel
func (i *requests) ListenFor(channel string) {
	// listen on a channel
	reqLock.Lock()
	reqChannels[channel] = append(reqChannels[channel], i)
	reqLock.Unlock()
	if i.wrapped != nil {
		i.wrapped.ListenFor(channel)
	}
}

// OnMessage assigns a callback function to be called when a message
// is received on this transport
func (i *requests) OnMessage(messageFunc qp.MessageFunc) {
	// assign the callback to be called
	i.callback = messageFunc
	if i.wrapped != nil {
		i.wrapped.OnMessage(messageFunc)
	}
}

// Send sends a message into the transport
func (i *requests) Send(channel string, message []byte) error {
	reqLock.RLock()
	_, ok := reqChannels[channel]
	reqLock.RUnlock()
	if ok {
		reqQueue <- &qp.Message{Source: channel, Data: message}
	} else {
		if i.wrapped != nil {
			return i.wrapped.Send(channel, message)
		}
	}
	return nil
}

// SetTimeout is a no-op for the InProc transport
func (i *requests) SetTimeout(timeout time.Duration) {
}

// Start is a no-op for the InProc transport.
func (i *requests) Start() error {
	return nil
}

// Stop is a no-op for the InProc transport.
func (i *requests) Stop() <-chan stop.Signal {
	return stop.Stopped()
}
