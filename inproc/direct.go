package inproc

import (
	"sync"
	"time"

	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

// Direct represents a qp.DirectTransport.
type Direct struct {
	lock     sync.RWMutex
	handlers map[string]qp.Handler
	stopChan chan stop.Signal
	Logger   qp.Logger
}

// ensure the interface is satisfied
var _ qp.DirectTransport = (*Direct)(nil)

var directQueue = make(chan *qp.Message)
var directInstances = make(map[*Direct]struct{})
var directLock sync.RWMutex

// NewDirect makes a new Direct.
func NewDirect() *Direct {
	p := &Direct{
		handlers: make(map[string]qp.Handler),
		Logger:   qp.NilLogger,
	}
	directLock.Lock()
	directInstances[p] = exists
	directLock.Unlock()
	return p
}

func processDirect() {
	go func() {
		for {
			select {
			case m, ok := <-directQueue:
				if !ok {
					return
				}
				directLock.Lock()
				for instance := range directInstances {
					instance.lock.RLock()
					if h, ok := instance.handlers[m.Source]; ok {
						go h.Handle(m)
						instance.lock.RUnlock()
						break
					}
					instance.lock.RUnlock()
				}
				directLock.Unlock()
			}
		}
	}()
}

func init() {
	processDirect()
}

// Send sends a message to the given chanenl
func (p *Direct) Send(channel string, data []byte) error {
	m := &qp.Message{Source: channel, Data: data}
	p.Logger.Infof("send %v", m)
	directQueue <- m
	return nil
}

// OnMessage binds the handler to the specified channel.
func (p *Direct) OnMessage(channel string, handler qp.Handler) error {
	p.lock.Lock()
	p.handlers[channel] = handler
	p.lock.Unlock()
	p.Logger.Info("OnMessage ", channel)
	return nil
}

// Start starts the transport.
func (p *Direct) Start() error {
	p.stopChan = stop.Make()
	p.Logger.Info("start inproc direct")
	return nil
}

// Stop stops the transport and closes StopChan() when finished.
func (p *Direct) Stop(time.Duration) {
	p.Logger.Info("stop inproc direct")
	directLock.Lock()
	delete(directInstances, p)
	directLock.Unlock()
	close(p.stopChan)
}

// StopChan gets the stop channel which will be closed when
// this transport has successfully stopped.
func (p *Direct) StopChan() <-chan stop.Signal {
	return p.stopChan
}
