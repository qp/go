package transports

// BinaryMessage is used to communicate both the
// topic of the message and the associated data.
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
	Send(to string, data []byte) error
	ListenFor(topic string)
	OnMessage(MessageFunc)
}
