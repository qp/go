package qp

import (
	"log"
	"sync"
)

// errResolving represents failure to resolve requests.
type errResolving struct {
	ID RequestID
}

// Error gets a string that describes this error.
func (e errResolving) Error() string {
	return "qp: Failed to resolve response " + string(e.ID)
}

// Request defines all the fields and information
// in the standard qp request object. It is used
// as part of the Handler callback.
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
func newRequest(endpoint string, object interface{}, pipeline []string) *Request {
	return &Request{To: pipeline, From: []string{endpoint}, ID: unique(), Data: object}
}

// Requester makes requests.
type Requester struct {
	name            string
	instanceID      string
	codec           Codec
	transport       DirectTransport
	responseChannel string
	resolver        *reqResolver
}

// NewRequester makes a new object capable of making requests and handling responses.
func NewRequester(name, instanceID string, codec Codec, transport DirectTransport) *Requester {
	r := &Requester{
		transport: transport,
		codec:     codec,
		resolver:  newResolver(),
	}
	r.responseChannel = name + "." + instanceID
	r.transport.OnMessage(r.responseChannel, HandlerFunc(func(m *Message) {
		var response Response
		if err := r.codec.Unmarshal(m.Data, &response); err != nil {
			log.Println("TODO: handle borked response", err)
			return
		}
		go func() {
			if err := r.resolver.Resolve(&response); err != nil {
				log.Println("TODO: handle error", err)
			}
		}()
	}))

	return r
}

// Issue issues the request and returns a Future from which you can
// get the response.
// The pipeline may be one or more endpoints. If it is more than one, each will receive
// the message, in order, and have an opportunity to mutate it before it is dispatched
// to the next endpoint in the pipeline.
// The provided object will be serialized and send as the "data" field in the message.
func (r *Requester) Issue(pipeline []string, obj interface{}) (*Future, error) {

	request := newRequest(r.responseChannel, obj, pipeline[1:])
	to := pipeline[0]
	bytes, err := r.codec.Marshal(request)
	if err != nil {
		return nil, err
	}
	f := newFuture(request.ID)
	r.resolver.Track(f)
	r.transport.Send(to, bytes)

	return f, nil
}

// Future implements a future for a response object
// It allows execution to continue until the response object
// is requested from this object, at which point it blocks and
// waits for the response to come back.
type Future struct {
	id       RequestID
	response chan *Response
	cached   *Response
	fetched  chan Signal
}

// newFuture creates a new response future that
// is initialized appropriately for waiting on an incoming
// response.
func newFuture(id RequestID) *Future {
	return &Future{id: id, response: make(chan *Response), fetched: make(chan Signal)}
}

// Response uses a future mechanism to retrieve the response.
// Execution continues asynchronously until this method is called,
// at which point execution blocks until the Response object is
// available.
//
// There is no timeout. It will block indefinitely. This may
// change in the future.
func (r *Future) Response() *Response {
	select {
	case <-r.fetched:
		return r.cached
	case r.cached = <-r.response:
		close(r.fetched)
		return r.cached
	}
}

// RequestResolver is responsible for tracking futures
// and resolving them when a response is received
type reqResolver struct {
	items map[RequestID]*Future
	lock  sync.Mutex
}

// newResolver creates and initializes a
// resolver object
func newResolver() *reqResolver {
	return &reqResolver{items: map[RequestID]*Future{}}
}

// Track begins tracking a Future, waiting for
// a response to come in
func (c *reqResolver) Track(future *Future) {
	c.lock.Lock()
	c.items[future.id] = future
	c.lock.Unlock()
}

// Resolve resolves a Future by matching it up
// with the given Response
func (c *reqResolver) Resolve(response *Response) error {
	var future *Future
	c.lock.Lock()
	future = c.items[response.ID]
	delete(c.items, response.ID)
	c.lock.Unlock()
	if future == nil {
		return &errResolving{ID: response.ID}
	}
	future.response <- response
	return nil
}

// Response defines all the fields and information
// included as part of a response to a request.
type Response struct {
	// From is an array of addresses encountered thus far
	From []string `json:"from"`
	// ID is the ID of the request to which this response relates
	ID RequestID `json:"id"`
	// Data is the repsonse data payload
	Data interface{} `json:"data"`
}
