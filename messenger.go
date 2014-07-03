package qp

import (
	"fmt"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/qp/go/codecs"
	"github.com/qp/go/messages"
	"github.com/qp/go/transports"
	"github.com/qp/go/utils"
)

// ListenFunc is the signature for the callback function called when
// a message is received. If an error occurs during processing,
// return an object describing the error. It will be assigned
// to the Err field in the standard message object and returned
// to the origin server of the message.
type ListenFunc func(message *messages.Message) interface{}

// Messenger is the entry point into the qp system.
// It is used to communicate with the underlying queue
// technology in an agnostic way, allowing for the codec and
// transport to be swapped out at any time.
type Messenger struct {
	name      string // the name of this service
	id        string // the UUID of this service
	fullName  string // the full name (name+UUID) of this service
	requests  map[string]*Response
	OnRequest ListenFunc
	codec     codecs.Codec
	transport transports.Transport
	lock      sync.RWMutex
}

// NewMessenger creates a new messenger that can be used to interact with the
// provided transport using the provided codec.
//
// Messenger automatically subscribes to two topics: "name" and "name/<UUID>". This
// allows for other services to send messages to this service, as well as respond directly
// to its unique ID.
//
// When a service-wide message is received as part of a pipeline, the "onRequest" callback is called.
// Do your work in the callback, mutating the data field of the message as necessary. If you
// encounter an error, return it and the system will abort the pipeline, returning your error back to
// the original sender.
func NewMessenger(name string, codec codecs.Codec, transport transports.Transport) *Messenger {
	id := uuid.New()
	m := &Messenger{
		name:      name,
		id:        id,
		fullName:  name + "/" + id,
		requests:  map[string]*Response{},
		codec:     codec,
		transport: transport,
	}

	m.transport.ListenFor(name, func(bm *transports.BinaryMessage) {
		msg := &messages.Message{}
		err := codec.Unmarshal(bm.Data, msg)
		if err != nil {
			// TODO: figure out how to handle this. can't unmarshal means
			// can't get the origin server. how to respond?
			return
		}

		if m.OnRequest != nil {
			msg.Err = m.OnRequest(msg)
		}
		next := ""
		if msg.Err != nil {
			next = msg.From[0]
		} else {
			next = msg.To.Pop()
			if next == "" {
				next = msg.From[0]
			}
		}
		msg.From.BPush(m.fullName)

		bytes, err := m.codec.Marshal(msg)
		if err != nil {
			// we can't marshal this msg object for some reason, so
			// we need to create a new one, set the to, and send what info we can
			failMsg := messages.NewMessage(m.fullName, nil, msg.From[0])
			failMsg.Err = map[string]interface{}{"message": "unable to marshal message object"}
			next = failMsg.To.Pop()
			failMsg.From.BPush(m.fullName)
			bytes, err = m.codec.Marshal(msg)
			if err != nil {
				// just give up
				return
			}
		}
		err = m.transport.Send(next, bytes)
		if err != nil {
			fmt.Println("send error: ", err)
			// TODO: figure out how to handle this. cannot send, so can't inform origin
			// or pass on to next endpoint
		}

	})

	m.transport.ListenFor(m.fullName, func(bm *transports.BinaryMessage) {
		msg := &messages.Message{}
		err := codec.Unmarshal(bm.Data, msg)
		if err != nil {
			fmt.Println("unmarshal error in direct: ", err)
			// TODO: pack this error into the error field of a new msg object
		}

		m.lock.Lock()
		r := m.requests[msg.ID]
		delete(m.requests, msg.ID)
		m.lock.Unlock()

		r.response <- msg
	})

	return m
}

// Request makes a new request to the given destination pipeline.
// The pipeline can be one or many destinations. If multiple arguments
// are given, the individual endpoints are visited, in order, and each
// has an opportunity to act on the request before forwarding it on.
// When the final destination is reached, it will send the response directly
// back to this service.
//
// The returned Response object provides a future through which you can wait on
// the response Message object.
func (m *Messenger) Request(object interface{}, pipeline ...string) (*Response, error) {
	stack := utils.StringDES(pipeline)
	to := stack.Pop()
	msg := messages.NewMessage(m.fullName, object, stack...)
	bytes, err := m.codec.Marshal(msg)
	if err != nil {
		return nil, err
	}

	err = m.transport.Send(to, bytes)
	if err != nil {
		return nil, err
	}

	r := newResponse(1 * time.Second)
	m.lock.Lock()
	m.requests[msg.ID] = r
	m.lock.Unlock()

	return r, nil

}

// Stop deregisters all listening topics
func (m *Messenger) Stop() {
	m.transport.Stop()
	m.lock.Lock()
	m.requests = nil
	m.lock.Unlock()
}

// Name returns the name of this messenger
func (m *Messenger) Name() string {
	return m.name
}

// ID returns the ID of this messenger
func (m *Messenger) ID() string {
	return m.id
}

// FullName returns the full name of this messenger
func (m *Messenger) FullName() string {
	return m.fullName
}
