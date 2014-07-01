package transport

import "github.com/qp/go/messages"

// BinaryMessage is used to communicate both the
// topic of the message and the associated data.
type BinaryMessage struct {
	topic string
	data  []byte
}

// MessageFunc is the signature for a Message Received Callback
type MessageFunc func(bm *BinaryMessage)

// Transport is an interface declaring functions used
// for interacting with an underlying transport technology
// such as nsq or rabbitmq.
type Transport interface {
	// listen for a message on the given topic
	// must be called before calling Start
	ListenFor(topic string)
	OnMessage(callback MessageFunc)       // set the function to be called when a message is received
	Send(message *messages.Message) error // send a message to the queue
	Start()                               // start processing messages
	Stop()                                // gracefully stop processing messages
}
