package qp

import (
	"log"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/qp/go/codecs"
	"github.com/qp/go/exchange"
	"github.com/qp/go/transports"
)

// EventMessenger allows for pub/sub communication in the qp
// syste. An EventMessenger has no concept of a response. All
// messages published are strictly fire and forget. Any interested
// listeners who are subscribed to the channel will receive
// the message.
type EventMessenger struct {
	name      string
	id        string
	codec     codecs.Codec
	transport transports.EventTransport
	mapper    *exchange.EventMapper
}

// MakeEventMessenger creates a new event messenger
// instanceName is a unique identifier for this particular instance of
// an endpoint. If it is empty, a unique ID will be generated for you.
func MakeEventMessenger(name, instanceName string, codec codecs.Codec, transport transports.EventTransport) *EventMessenger {
	if instanceName == "" {
		instanceName = uuid.New()
	}
	e := &EventMessenger{
		name:      name,
		id:        name + "." + instanceName,
		codec:     codec,
		transport: transport,
		mapper:    exchange.MakeEventMapper(),
	}

	e.transport.OnMessage(func(bm *transports.BinaryMessage) {
		if cbs := e.mapper.Find(bm.Channel); cbs != nil {

			// decode to event object
			var event exchange.Event
			err := e.codec.Unmarshal(bm.Data, &event)
			if err != nil {
				// dispatch a log entry and abort
				log.Println("Unable to unmarshal event: ", err)
				return
			}

			for _, cb := range cbs {
				cb(bm.Channel, &event)
			}
		}
	})

	return e
}

// SetTimeout sets the timeout to the given value.
// This timeout is used when gracefully shutting down the
// transport. In-flight requests will have this much time
// to complete before being abandoned.
// The default timeout value is 5 seconds.
func (e *EventMessenger) SetTimeout(timeout time.Duration) {
	e.transport.SetTimeout(timeout)
}

// Start spins up the messenger to begin processing messages
func (e *EventMessenger) Start() {
	e.transport.Start()
}

// Stop spinds down the messenger gracefully, allowing in-flight
// messages to be processed, but accepting no new ones
func (e *EventMessenger) Stop() {
	e.transport.Stop()
}

// Subscribe registers a handler for a list of event channels. When an event is
// received on any of the given channels, the provided handler is called.
//
// A channel can take the form of `prefix[.optional_prefix(es)][.*]`
//
// For example, `router.request.type` could be used to publish events about the
// types of requests being handled by a router. `router.request.size` could be used
// to publish events about the size of requests being handled by a router. At this point,
// you could subscribe to `router.request.*` and receive both type and size messages.
func (e *EventMessenger) Subscribe(handler exchange.EventHandler, channels ...string) {
	// validate handler is not nil
	if handler == nil {
		panic("handler cannot be nil")
	}
	// validate channels is not empty
	if len(channels) == 0 {
		panic("channels cannot be empty")
	}

	// associate each channel with the appropriate handler function
	for _, channel := range channels {
		e.mapper.Track(channel, handler)
		// instruct the transport to listen on the channel
		e.transport.ListenFor(channel)
	}
}

// SubscribeChildren registers a handler for a list of event channels and their children.
// When an event is received on any of the given channels, the provided handler
// is called.
//
// A channel can take the form of `prefix[.optional_prefix(es)][.*]`
//
// For example, `router.request.type` could be used to publish events about the
// types of requests being handled by a router. `router.request.size` could be used
// to publish events about the size of requests being handled by a router. At this point,
// you could subscribe to `router.request.*` and receive both type and size messages.
func (e *EventMessenger) SubscribeChildren(handler exchange.EventHandler, channels ...string) {
	// validate handler is not nil
	if handler == nil {
		panic("handler cannot be nil")
	}
	// validate channels is not empty
	if len(channels) == 0 {
		panic("channels cannot be empty")
	}

	// associate each channel with the appropriate handler function
	for _, channel := range channels {
		e.mapper.Track(channel, handler)
		// instruct the transport to listen on the channel
		e.transport.ListenForChildren(channel)
	}
}

// Send publishes an event, encapsulating the given object so user data
// can be sent along with it.
func (e *EventMessenger) Send(object interface{}, channel string) error {
	// validate that we have a channels
	if len(channel) == 0 {
		panic("channel cannot be empty")
	}
	//validate that we have an object
	if object == nil {
		panic("object cannot be nil")
	}

	// make a new event object and give it the service unique ID and the object to send
	event := exchange.MakeEvent(e.id, object)

	// encode the event object to a byte slice
	data, err := e.codec.Marshal(event)
	if err != nil {
		return err
	}
	return e.transport.Send(channel, data)
}
