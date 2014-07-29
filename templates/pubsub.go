package templates

import (
	"errors"
	"time"

	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

// PubSub represents a qp.PubSubTransport.
type PubSub struct {
	stopChan chan stop.Signal
}

// ensure the interface is satisfied
var _ qp.PubSubTransport = (*PubSub)(nil)

// NewPubSub makes a new PubSub.
func NewPubSub() *PubSub {
	p := &PubSub{
		stopChan: stop.Make(),
	}
	return p
}

// Publish publishes data on the specified channel.
func (p *PubSub) Publish(channel string, data []byte) error {
	return errors.New("not implemented")
}

// Subscribe binds the handler to the specified channel.
func (p *PubSub) Subscribe(channel string, handler qp.Handler) error {
	return errors.New("not implemented")
}

// Start starts the transport.
func (p *PubSub) Start() error {
	return errors.New("not implemented")
}

// Stop stops the transport and closes StopChan() when finished.
func (p *PubSub) Stop(grace time.Duration) {
	// do work to stop
	close(p.stopChan)
}

// StopChan gets the stop channel which will be closed when
// this transport has successfully stopped.
func (p *PubSub) StopChan() <-chan stop.Signal {
	return p.stopChan
}
