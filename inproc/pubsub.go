package inproc

import (
	"sync"
	"time"

	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

var nothing = struct{}{}

// PubSub represents a qp.PubSubTransport.
type PubSub struct {
	lock     sync.RWMutex
	handlers map[string]qp.Handler
	stopChan chan stop.Signal
}

// ensure the interface is satisfied
var _ qp.PubSubTransport = (*PubSub)(nil)

var queue = make(chan *qp.Message)
var instances = make(map[*PubSub]struct{})
var lock sync.RWMutex

// NewPubSub makes a new PubSub.
func NewPubSub() *PubSub {
	p := &PubSub{
		handlers: make(map[string]qp.Handler),
		stopChan: stop.Make(),
	}
	lock.Lock()
	instances[p] = nothing
	lock.Unlock()
	return p
}

func processMessages() {
	go func() {
		for {
			select {
			case m, ok := <-queue:
				if !ok {
					return
				}
				lock.Lock()
				for instance := range instances {
					instance.lock.RLock()
					if h, ok := instance.handlers[m.Source]; ok {
						go h.Handle(m)
					}
					instance.lock.RUnlock()
				}
				lock.Unlock()
			}
		}
	}()
}

func init() {
	processMessages()
}

// Publish publishes data on the specified channel.
func (p *PubSub) Publish(channel string, data []byte) error {
	m := &qp.Message{Source: channel, Data: data}
	queue <- m
	return nil
}

// Subscribe binds the handler to the specified channel.
func (p *PubSub) Subscribe(channel string, handler qp.Handler) error {
	p.lock.Lock()
	p.handlers[channel] = handler
	p.lock.Unlock()
	return nil
}

// Start starts the transport.
func (p *PubSub) Start() error {
	return nil
}

// Stop stops the transport and closes StopChan() when finished.
func (p *PubSub) Stop(time.Duration) {
	lock.Lock()
	delete(instances, p)
	lock.Unlock()
	close(p.stopChan)
}

// StopChan gets the stop channel which will be closed when
// this transport has successfully stopped.
func (p *PubSub) StopChan() <-chan stop.Signal {
	return p.stopChan
}
