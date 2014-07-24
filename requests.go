package qp

import (
	"log"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

// RequestID represents a unique ID for a Request.
type RequestID uint64

// Request defines all the fields and information
// in the standard qp request object. It is used
// as part of the RequestHandler callback.
type Request struct {
	// To is an array of destination addresses
	To []string `json:"to"`
	// From is an array of addresses encountered thus far
	From []string `json:"from"`
	// ID is a number identifying this message
	ID RequestID `json:"id"`
	// Data is an arbitrary data payload
	Data interface{} `json:"data"`
}

// newRequest makes a new request object and generates a unique ID in the from array.
func newRequest(endpoint string, object interface{}, pipeline ...string) *Request {
	return &Request{To: pipeline, From: []string{endpoint}, ID: unique(), Data: object}
}

// Requester defines the interface through which
// requests are introduced into the qp system, and responses
// to those requests are returned.
type Requester struct {
	name         string
	responseName string
	codec        Codec
	transport    RequestTransport
	resolver     *reqResolver
	mapper       *reqMapper
}

// NewRequester creates a new request messenger that allows direct communication between
// two endpoints in the qp system. It also allows for pipelining through multiple specified
// endpoints.
func NewRequester(name, instanceName string, codec Codec, transport RequestTransport) *Requester {
	if instanceName == "" {
		instanceName = uuid.New()
	}

	r := &Requester{name: name,
		responseName: name + "." + instanceName,
		codec:        codec,
		transport:    transport,
		resolver:     newResolver(),
		mapper:       newReqMapper(),
	}

	// listen on the "responseName" channel
	r.transport.ListenFor(r.responseName)

	r.transport.OnMessage(func(bm *Message) {
		// switch on the bm.channel to determine the type of message
		if bm.Source == r.responseName {
			// decode to response object
			var response Response
			err := r.codec.Unmarshal(bm.Data, &response)
			if err != nil {
				// dispatch a log entry and abort
				log.Println("Unable to unmarshal response: ", err)
				return
			}

			// map the response to the appropriate ResponseFuture
			go r.resolver.Resolve(&response)
		} else {
			// decode to request object
			var Request Request
			err := r.codec.Unmarshal(bm.Data, &Request)
			if err != nil {
				log.Println("Unable to unmarshal request: ", err)
				return
			}

			// map the request to the appropriate RequestHandler
			handlers := r.mapper.Find(bm.Source)
			if handlers != nil {
				go func() {
					for _, handler := range handlers {
						// call each RequestHandler
						handler(bm.Source, &Request)
					}

					// get the next destination endpoint
					to := ""
					if len(Request.To) != 0 {
						to = Request.To[0]
						Request.To = Request.To[1:]
					} else {
						to = Request.From[0]
					}
					Request.From = append(Request.From, r.responseName)

					// encode the request
					data, err := r.codec.Marshal(&Request)
					if err != nil {
						log.Println("Unable to marshal request to send to next endpoint: ", err)
						return
					}

					// send the request to the next endpoint in the pipeline
					err = r.transport.Send(to, data)
					if err != nil {
						log.Println("Unable to send request to next endpoint: ", err)
						return
					}
				}()
			}
		}
	})

	return r
}

// SetTimeout sets the timeout to the given value.
// This timeout is used when gracefully shutting down the
// transport. In-flight requests will have this much time
// to complete before being abandoned.
// The default timeout value is 5 seconds.
func (r *Requester) SetTimeout(timeout time.Duration) {
	r.transport.SetTimeout(timeout)
}

// Start spins up the Request Messenger to start processing
// incoming and outgoing messages.
func (r *Requester) Start() error {
	// spin up the underlying transport
	return r.transport.Start()
}

// Stop tears down the Request Messenger in a graceful manner, allowing
// any in-flight operations to complete.
// After calling Stop, you may call Start again to resume processing. However,
// typically this method is called only once.
func (r *Requester) Stop() {
	// spin down the underlying transport
	r.transport.Stop()
}

// OnRequest assigns the given handler to the given channels, calling the handler
// when a message is received on one of those channels.
func (r *Requester) OnRequest(handler RequestHandler, channels ...string) {

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
		r.mapper.Track(channel, handler)
		// instruct the transport to listen on the channel
		r.transport.ListenFor(channel)
	}

}

// Request constructs a request to be sent to the given pipeline endpoint(s). The pipeline may
// be one or more endpoints. If it is more than one, each will receive the message, in order, and
// have an opportunity to mutate it before it is dispatched to the next endpoint in the pipeline.
// The provided object will be serialized and send as the "data" field in the message.
func (r *Requester) Request(object interface{}, pipeline ...string) (*ResponseFuture, error) {

	// validate that we have a pipeline
	if len(pipeline) == 0 {
		panic("pipeline cannot be empty")
	}
	//validate that we have an object
	if object == nil {
		panic("object cannot be nil")
	}

	// create a new request message object
	// assign the given "object" to the "data" field in the request object
	// assign the "to" stack in the request object using the pipeline string, except the
	// top that has been poppped off and is being used to make the transport call
	Request := newRequest(r.responseName, object, pipeline[1:]...)
	to := pipeline[0]

	// encode the request object to a byte slice
	data, err := r.codec.Marshal(Request)
	if err != nil {
		return nil, err
	}

	// use the unique ID in the request object to associate a request with a reply
	// we have to map the request to the response future, then handle that response when it comes back
	rf := newResponseFuture(Request.ID)
	r.resolver.Track(rf)

	// call the transport and give it the popped "to" endpoint, as well as
	// the request object to that endpoint and give it the encoded message
	err = r.transport.Send(to, data)
	if err != nil {
		return nil, err
	}

	return rf, nil
}

// RequestHandler defines the function signature for the callback
// that will be called when a request is received.
type RequestHandler func(channel string, Request *Request)

// reqMapper uses a map internally to implement
// the mapper interface
type reqMapper struct {
	lock  sync.RWMutex
	items map[string][]RequestHandler
}

// newReqMapper initializes and returns a mapper instance
// as a mapper interface.
func newReqMapper() *reqMapper {
	return &reqMapper{items: map[string][]RequestHandler{}}
}

// Track begins tracking an id and its associated handler so it
// can be found later
func (m *reqMapper) Track(id string, handler RequestHandler) {
	m.lock.Lock()
	m.items[id] = append(m.items[id], handler)
	m.lock.Unlock()

}

// Find locates the given id and returns the handlers associated with it
func (m *reqMapper) Find(id string) []RequestHandler {
	var handlers []RequestHandler
	m.lock.RLock()
	handlers = m.items[id]
	m.lock.RUnlock()
	return handlers
}

// RequestResolver is responsible for tracking futures
// and resolving them when a response is received
type reqResolver struct {
	items map[RequestID]*ResponseFuture
	lock  sync.Mutex
}

// newResolver creates and initializes a
// resolver object
func newResolver() *reqResolver {
	return &reqResolver{items: map[RequestID]*ResponseFuture{}}
}

// Track begins tracking a ResponseFuture, waiting for
// a response to come in
func (c *reqResolver) Track(future *ResponseFuture) {
	c.lock.Lock()
	c.items[future.id] = future
	c.lock.Unlock()
}

// Resolve resolves a ResponseFuture by matching it up
// with the given Response
func (c *reqResolver) Resolve(response *Response) {
	var future *ResponseFuture
	c.lock.Lock()
	future = c.items[response.ID]
	delete(c.items, response.ID)
	c.lock.Unlock()
	future.response <- response
}
