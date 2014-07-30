package qp

import "log"

// Event defines all the fields and information
// included as part of a Event to a request.
type Event struct {
	// From is the address of the sender.
	From string `json:"from"`
	// Data is the payload of the event.
	Data interface{} `json:"data"`
}

// Publisher represents types capable of publishing events.
type Publisher interface {
	// Publish publishes the object on the specified channel.
	Publish(channel string, obj interface{}) error
}

// publisher allows events to be published.
type publisher struct {
	name       string
	instanceID string
	uniqueID   string
	codec      Codec
	transport  PubSubTransport
}

// NewPublisher makes a new publisher capable of Publishing events.
func NewPublisher(name, instanceID string, codec Codec, transport PubSubTransport) Publisher {
	return &publisher{
		name:       name,
		instanceID: instanceID,
		uniqueID:   name + "." + instanceID,
		codec:      codec,
		transport:  transport,
	}
}

func (p *publisher) Publish(channel string, obj interface{}) error {

	event := &Event{From: p.uniqueID, Data: obj}
	data, err := p.codec.Marshal(event)
	if err != nil {
		return err
	}
	if err := p.transport.Publish(channel, data); err != nil {
		return err
	}
	return nil

}

// EventHandler represents types capable of handling Requests.
type EventHandler interface {
	Handle(*Event)
}

// EventHandlerFunc represents functions capable of handling
// Requests.
type EventHandlerFunc func(*Event)

// Handle calls the EventHandlerFunc in order to handle
// the specific Event.
func (f EventHandlerFunc) Handle(r *Event) {
	f(r)
}

// Subscriber represents types capable of subscribing to
// events.
type Subscriber interface {
	// Subscribe binds the handler to the specified channel.
	Subscribe(channel string, handler EventHandler) error
	// SubscribeFunc binds the EventHandlerFunc to the specified channel.
	SubscribeFunc(channel string, fn EventHandlerFunc) error
}

// subscriber allows events to be subscribed to.
type subscriber struct {
	codec     Codec
	transport PubSubTransport
}

// NewSubscriber creates a Subscriber object capable of subscribing
// to events.
func NewSubscriber(codec Codec, transport PubSubTransport) Subscriber {
	return &subscriber{codec: codec, transport: transport}
}

func (s *subscriber) Subscribe(channel string, handler EventHandler) error {
	return s.transport.Subscribe(channel, HandlerFunc(func(msg *Message) {

		var event Event
		if err := s.codec.Unmarshal(msg.Data, &event); err != nil {
			log.Println("TODO: Handle unmsrshal error in Subscribe:", err)
			return
		}

		handler.Handle(&event)

	}))
}

func (s *subscriber) SubscribeFunc(channel string, fn EventHandlerFunc) error {
	return s.Subscribe(channel, fn)
}
