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

// transport is an interface declaring functions used
// for interacting with an underlying transport technology
// such as nsq or rabbitmq.
type transport interface {
	Start() error
	Stop()
	OnMessage(messageFunc MessageFunc)
	SetTimeout(timeout time.Duration)
	ListenFor(channel string)
	Send(to string, data []byte) error
}

// RequestTransport describes types capable of making Requests and
// getting responses.
type RequestTransport interface {
	transport
}

// EventTransport describes types capable of firing and listening
// for events.
type EventTransport interface {
	transport
	ListenForChildren(channel string)
}
