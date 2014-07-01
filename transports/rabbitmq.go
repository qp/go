package transport

import "github.com/qp/go/shared"

// ensure RabbitMQ conforms to Transport interface
var conformsRabbitMQ = Transport(&RabbitMQ{})

// RabbitMQ is the RabbitMQ implementation of the
// Transport interface. It provides all functionality
// necessary to fulfill the Transport contract through
// a RabbitMQ transport layer.
type RabbitMQ struct {
}

// ListenFor instructs RabbitMQ to deliver a message for the given topic
func (r *RabbitMQ) ListenFor(topic string) {

}

// OnMessage sets the callback function to call when a message is received
func (r *RabbitMQ) OnMessage(callback MessageFunc) {

}

// Send sends a message out to RabbitMQ
func (r *RabbitMQ) Send(message *shared.Message) error {
	return nil
}

// Start begins processing messages to/from RabbitMQ
func (r *RabbitMQ) Start() {

}

// Stop gracefully stops processing messages, allowing in-flight
// requests to finish before stopping entirely
func (r *RabbitMQ) Stop() {

}
