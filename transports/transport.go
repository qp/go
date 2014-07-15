package transports

import (
	"errors"
	"time"
)

// ErrTransportStopped is returned when an method is
// called on a stopped transport
var ErrTransportStopped = errors.New("transport is stopped")

// BinaryMessage is used to communicate both the
// channel of the message and the associated data.
type BinaryMessage struct {
	Channel string
	Data    []byte
}

// MessageFunc is the signature for a Message Received Callback
type MessageFunc func(bm *BinaryMessage)

// Transport is an interface declaring functions used
// for interacting with an underlying transport technology
// such as nsq or rabbitmq.
type Transport interface {
	Start() error
	Stop()
	OnMessage(messageFunc MessageFunc)
	SetTimeout(timeout time.Duration)
}

// RequestTransport extends Transport to provide a way to
// send messages to a given endpoint.
type RequestTransport interface {
	Transport
	ListenFor(channel string)
	Send(to string, data []byte) error
}

// EventTransport extends Transport to provide a way to
// publish messages to subscribed listeners.
type EventTransport interface {
	Transport
	ListenFor(channel string)
	ListenForChildren(channel string)
	Publish(to string, data []byte) error
}
