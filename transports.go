package qp

import (
	"errors"

	"github.com/stretchr/pat/start"
)

// ErrNotRunning is returned when an method is
// called on a transport that is not running.
var ErrNotRunning = errors.New("transport is not running")

// ErrRunning is returned when an method is
// called on a transport that is running.
var ErrRunning = errors.New("transport is running")

// Message represents a single message of data and its source.
type Message struct {
	// The channel the Message came from.
	Source string
	// The data of the message.
	Data []byte
}

// Handler represents types capable of handling messages
// from the transports.
type Handler interface {
	Handle(msg *Message)
}

// HandlerFunc represents functions capable of handling
// messages.
type HandlerFunc func(msg *Message)

// Handle calls the HandlerFunc.
func (f HandlerFunc) Handle(msg *Message) {
	f(msg)
}

// PubSubTransport represents a transport capable of
// providing publish/subscribe capabilities.
type PubSubTransport interface {
	start.StartStopper
	// Publish publishes data on the specified channel.
	Publish(channel string, data []byte) error
	// Subscribe binds the handler to the specified channel.
	// Only one handler can be associated with a given channel.
	// Multiple calls to Subscribe with the same channel will replace the previous handler.
	Subscribe(channel string, handler Handler) error
}

// DirectTransport represents a transport capable of
// providing request/response capabilities.
type DirectTransport interface {
	start.StartStopper
	// Send sends data on the channel.
	Send(channel string, data []byte) error
	// OnMessage binds the handler to the specified channel.
	// Only one handler can be associated with a given channel.
	// Multiple calls to OnMessage wiht the same channel will replace the previous handler.
	OnMessage(channel string, handler Handler) error
}
