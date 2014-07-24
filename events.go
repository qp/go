package qp

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

// Event defines all the fields and information
// included as part of a Event to a request.
type Event struct {
	// From is the address of the sender.
	From string `json:"from"`
	// Data is the payload of the event.
	Data interface{} `json:"data"`
}

// newEvent makes a new Event object
func newEvent(endpoint string, object interface{}) *Event {
	return &Event{From: endpoint, Data: object}
}

// String gets a string representation of this Event.
func (r Event) String() string {
	return fmt.Sprintf("From: %v\nData: %v", r.From, r.Data)
}

// EventHandler defines the function signature for the callback
// that will be called when a event is received.
type EventHandler func(channel string, event *Event)

// eventMapper uses a map internally to implement
// the mapper interface
type eventMapper struct {
	lock  sync.RWMutex
	items map[string][]EventHandler
}

// newEventMapper initializes and returns a mapper instance
// as a mapper interface.
func newEventMapper() *eventMapper {
	return &eventMapper{items: map[string][]EventHandler{}}
}

// Track begins tracking an id and its associated handler so it
// can be found later
func (m *eventMapper) Track(id string, handler EventHandler) {
	m.lock.Lock()
	m.items[id] = append(m.items[id], handler)
	m.lock.Unlock()

}

// Find locates the given id and returns the handlers associated with it
func (m *eventMapper) Find(id string) []EventHandler {
	var handlers []EventHandler
	m.lock.RLock()
	for itemID, item := range m.items {
		if strings.HasSuffix(itemID, "*") {
			if strings.HasPrefix(id, itemID[:len(itemID)-1]) {
				handlers = append(handlers, item...)
			}
		} else if id == itemID {
			handlers = append(handlers, item...)
		}
	}
	m.lock.RUnlock()
	return handlers
}

// PubSub allows for pub/sub communication in the qp
// syste. An EventMessenger has no concept of a response. All
// messages published are strictly fire and forget. Any interested
// listeners who are subscribed to the channel will receive
// the message.
type PubSub struct {
	name      string
	id        string
	codec     Codec
	transport PubSubTransport
	mapper    *eventMapper
}

// NewPubSub creates a new PubSub for publishing and subscribing to events.
// instanceName is a unique identifier for this particular instance of
// an endpoint. If it is empty, a unique ID will be generated for you.
func NewPubSub(name, instanceName string, codec Codec, transport PubSubTransport) *PubSub {
	if instanceName == "" {
		instanceName = uuid.New()
	}
	e := &PubSub{
		name:      name,
		id:        name + "." + instanceName,
		codec:     codec,
		transport: transport,
		mapper:    newEventMapper(),
	}

	e.transport.OnMessage(func(bm *Message) {
		if cbs := e.mapper.Find(bm.Source); cbs != nil {
			// decode to event object
			var event Event
			err := e.codec.Unmarshal(bm.Data, &event)
			if err != nil {
				// dispatch a log entry and abort
				log.Println("Unable to unmarshal event: ", err)
				return
			}

			for _, cb := range cbs {
				cb(bm.Source, &event)
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
func (e *PubSub) SetTimeout(timeout time.Duration) {
	e.transport.SetTimeout(timeout)
}

// Start spins up the messenger to begin processing messages
func (e *PubSub) Start() error {
	return e.transport.Start()
}

// Stop spinds down the messenger gracefully, allowing in-flight
// messages to be processed, but accepting no new ones
func (e *PubSub) Stop() {
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
func (e *PubSub) Subscribe(handler EventHandler, channels ...string) {
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
func (e *PubSub) SubscribeChildren(handler EventHandler, channels ...string) {
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
		e.mapper.Track(channel+"*", handler)
		// instruct the transport to listen on the channel
		e.transport.ListenForChildren(channel)
	}
}

// Send publishes an event, encapsulating the given object so user data
// can be sent along with it.
func (e *PubSub) Send(object interface{}, channel string) error {
	// validate that we have a channels
	if len(channel) == 0 {
		panic("channel cannot be empty")
	}
	//validate that we have an object
	if object == nil {
		panic("object cannot be nil")
	}

	// make a new event object and give it the service unique ID and the object to send
	event := newEvent(e.id, object)

	// encode the event object to a byte slice
	data, err := e.codec.Marshal(event)
	if err != nil {
		return err
	}
	return e.transport.Send(channel, data)
}
