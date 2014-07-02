package transport

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
	// listen for a message on the given topic and
	// calls the given callback function when a message is
	// received. The callback function is fired in its own
	// goroutine to minimize latency at the transport level.
	ListenFor(topic string, callback MessageFunc) error
	Send(topic string, message []byte) error // send a message to the queue
	// Start processing messages. Dial the underlying transport if necessary.
	Start() error
	Stop() // gracefully stop processing messages
}
