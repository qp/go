package qp

import (
	"errors"
	"time"
)

type Signal struct{}

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

type Handler interface {
	Handle(msg *Message)
}

// HandlerFunc is the signature for a Message Received Callback.
type HandlerFunc func(msg *Message)

func (f HandlerFunc) Handle(msg *Message) {
	f(msg)
}

type PubSubTransport interface {
	Start() error
	Stop()
	StopChan() <-chan Signal
	Publish(channel string, data interface{})
	Subscribe(channel string, handler Handler)
}

type SenderTransport interface {
	Start() error
	Stop(wait time.Duration)
	StopChan() <-chan Signal
	Send(channel string, data interface{})
	OnMessage(channel string, handler Handler)
}
