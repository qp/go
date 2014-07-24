package inproc

import (
	"strings"
	"sync"
	"time"

	"github.com/qp/go"
)

// events is the InProc implementation of the
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
type events struct {
	callback qp.MessageFunc
	wrapped  qp.EventTransport
}

var evtQueue = make(chan *qp.Message)
var evtChannels = map[string][]*events{}
var evtLock sync.RWMutex

func processEventsMessages() {
	go func() {
		for {
			select {
			case bm := <-evtQueue:
				evtLock.RLock()
				for channel, procs := range evtChannels {
					if strings.HasSuffix(channel, "*") {
						if strings.HasPrefix(bm.Source, channel[:len(channel)-1]) {
							for _, instance := range procs {
								go instance.callback(bm)
							}
						}
					} else if channel == bm.Source {
						for _, instance := range procs {
							go instance.callback(bm)
						}
					}
				}
				evtLock.RUnlock()
			}
		}
	}()
}

func init() {
	processEventsMessages()
}

// NewPubSubTransport creates a new in-process qp.EventTransport.
func NewPubSubTransport(wrapped qp.EventTransport) qp.EventTransport {
	return &events{wrapped: wrapped}
}

// ListenFor instructs InProc to listen for a message for the given channel
func (i *events) ListenFor(channel string) {
	// listen on a channel
	evtLock.Lock()
	evtChannels[channel] = append(evtChannels[channel], i)
	evtLock.Unlock()
	if i.wrapped != nil {
		i.wrapped.ListenFor(channel)
	}
}

// ListenForChildren instructs InProc to listen for a message for the given channel
// and all its children
func (i *events) ListenForChildren(channel string) {
	channel += "*"
	// listen on a channel
	evtLock.Lock()
	evtChannels[channel] = append(evtChannels[channel], i)
	evtLock.Unlock()
	if i.wrapped != nil {
		i.wrapped.ListenFor(channel)
	}
}

// OnMessage assigns a callback function to be called when a message
// is received on this transport
func (i *events) OnMessage(messageFunc qp.MessageFunc) {
	// assign the callback to be called
	i.callback = messageFunc
	if i.wrapped != nil {
		i.wrapped.OnMessage(messageFunc)
	}
}

// Send sends a message into the transport
func (i *events) Send(channel string, message []byte) error {
	evtLock.RLock()
	_, ok := evtChannels[channel]
	evtLock.RUnlock()
	if ok {
		evtQueue <- &qp.Message{Source: channel, Data: message}
	} else {
		if i.wrapped != nil {
			return i.wrapped.Send(channel, message)
		}
	}
	return nil
}

// Start is a no-op for the InProc transport.
func (i *events) Start() error {
	return nil
}

// Stop is a no-op for the InProc transport.
func (i *events) Stop() {
}

// SetTimeout is a no-op for the InProc transport
func (i *events) SetTimeout(timeout time.Duration) {
}
