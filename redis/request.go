package redis

import (
	"time"

	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

type requests struct{}

var _ qp.RequestTransport = (*requests)(nil)

func NewReqTransport(url string) qp.RequestTransport {
	r := &requests{}
	return r
}

// Starts the transport.
func (r *requests) Start() error {
	return nil
}

func (r *requests) Stop() <-chan stop.Signal {
	return stop.Stopped()
}

// Registers a MessageFunc the will be called when a
// message is received.
func (r *requests) OnMessage(messageFunc qp.MessageFunc) {}

// SetTimeout sets the amount of time in-flight requests have
// to complete before being shut down.
func (r *requests) SetTimeout(timeout time.Duration) {}

// ListenFor listens for messages on the specified channel.
func (r *requests) ListenFor(channel string) {}

// Send sends a message of data to the specified destination.
func (r *requests) Send(to string, data []byte) error {
	return nil
}
