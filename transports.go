package qp

import (
	"errors"
	"time"
)

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

// MessageFunc is the signature for a Message Received Callback.
type MessageFunc func(bm *Message)

// RequestTransport describes types capable of making Requests and
// getting responses.
type RequestTransport interface {
	// Starts the transport.
	Start() error
	// Stops the transport.
	Stop()
	// Registers a MessageFunc the will be called when a
	// message is received.
	OnMessage(messageFunc MessageFunc)
	// SetTimeout sets the amount of time in-flight requests have
	// to complete before being shut down.
	SetTimeout(timeout time.Duration)
	// ListenFor listens for messages on the specified channel.
	ListenFor(channel string)
	// Send sends a message of data to the specified destination.
	Send(to string, data []byte) error
}

// PubSubTransport describes types capable of firing and listening
// for events.
type PubSubTransport interface {
	// Starts the transport.
	Start() error
	// Stops the transport.
	Stop()
	// Registers a MessageFunc the will be called when a
	// message is received.
	OnMessage(messageFunc MessageFunc)
	// SetTimeout sets the amount of time in-flight requests have
	// to complete before being shut down.
	SetTimeout(timeout time.Duration)
	// ListenFor listens for messages on the specified channel.
	ListenFor(channel string)
	// Send sends a message of data to the specified destination.
	Send(to string, data []byte) error
	// ListenForChildren listens for messages on the specified channel
	// or any of its children. See PubSub.SubscribeChildren for more
	// information.
	ListenForChildren(channel string)
}
