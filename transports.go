package qp

import "errors"

// ErrTransportStopped is returned when an method is
// called on a stopped transport.
var ErrTransportStopped = errors.New("transport is stopped")

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
	StartStopper
	Publish(channel string, data []byte) error
	Subscribe(channel string, handler Handler) error
}

// DirectTransport represents a transport capable of
// providing request/response capabilities.
type DirectTransport interface {
	StartStopper
	Send(channel string, data []byte) error
	OnMessage(channel string, handler Handler) error
}
