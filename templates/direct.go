package templates

import (
	"errors"
	"time"

	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

// Direct represents a qp.DirectTransport.
type Direct struct {
	stopChan chan stop.Signal
}

// ensure the interface is satisfied
var _ qp.DirectTransport = (*Direct)(nil)

// NewDirect makes a new direct transport.
func NewDirect() *Direct {
	return &Direct{
		stopChan: stop.Make(),
	}
}

// Send sends data on the channel.
func (d *Direct) Send(channel string, data []byte) error {
	return errors.New("not implemented")
}

// OnMessage binds the handler to the specified channel.
func (d *Direct) OnMessage(channel string, handler qp.Handler) error {
	return errors.New("not implemented")
}

// Start starts the transport.
func (d *Direct) Start() error {
	return errors.New("not implemented")
}

// Stop instructs the transport to gracefully stop and close the
// StopChan when stopping has completed.
//
// In-flight requests will have "wait" duration to complete
// before being abandoned.
func (d *Direct) Stop(wait time.Duration) {

}

// StopChan gets the stop channel which will block until
// stopping has completed, at which point it is closed.
// Callers should never close the stop channel.
func (d *Direct) StopChan() <-chan stop.Signal {
	return stop.Stopped()
}
