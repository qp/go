package qp

import (
	"errors"
	"sync"
	"time"

	"github.com/stretchr/slog"
)

// errResolving represents failure to resolve requests.
type errResolving struct {
	ID RequestID
}

// Error gets a string that describes this error.
func (e errResolving) Error() string {
	return "failed to resolve response " + string(e.ID)
}

// ErrTimeout represents situations when timeouts have occurred.
var ErrTimeout = errors.New("timed out")

// Transaction defines all the fields and information
// in the standard qp request object.
type Transaction struct {
	// To is an array of destination addresses
	To []string `json:"to"`
	// From is an array of addresses encountered thus far
	From []string `json:"from"`
	// ID is a number identifying this message
	ID RequestID `json:"id"`
	// Data is an arbitrary data payload
	Data interface{} `json:"data"`
}

// Abort clears the To slice indicating that the Transaction should
// be sent back to the originator.
func (r *Transaction) Abort() {
	r.To = []string{}
}

// newRequest makes a new request object and generates a unique ID in the from array.
func newTransaction(endpoint string, object interface{}, pipeline []string) *Transaction {
	return &Transaction{To: pipeline, From: []string{endpoint}, ID: unique(), Data: object}
}

// Requester represents a type capable of issuing requests and getting
// responses.
type Requester interface {
	// Issue issues the request and returns a Future from which you can
	// get the response.
	// The pipeline may be one or more endpoints. If it is more than one, each will receive
	// the message, in order, and have an opportunity to mutate it before it is dispatched
	// to the next endpoint in the pipeline.
	// The provided object will be serialized and send as the "data" field in the message.
	Issue(pipeline []string, obj interface{}) (*Future, error)
}

// Requester makes requests.
type requester struct {
	name            string
	instanceID      string
	codec           Codec
	transport       DirectTransport
	responseChannel string
	resolver        *reqResolver
	logger          slog.Logger
}

// NewRequester makes a new object capable of making requests and handling responses.
func NewRequester(name, instanceID string, codec Codec, transport DirectTransport) (Requester, error) {
	return NewRequesterLogger(name, instanceID, codec, transport, slog.NilLogger)
}

// NewRequesterLogger makes a new object capable of making requests and handling responses
// with logs going to the specified Logger.
func NewRequesterLogger(name, instanceID string, codec Codec, transport DirectTransport, logger slog.Logger) (Requester, error) {
	r := &requester{
		transport: transport,
		codec:     codec,
		resolver:  newResolver(),
		logger:    logger,
	}
	r.responseChannel = name + "." + instanceID

	err := r.transport.OnMessage(r.responseChannel, HandlerFunc(func(m *Message) {
		r.logger.Info("received on", r.responseChannel, m)
		var response Transaction
		if err := r.codec.Unmarshal(m.Data, &response); err != nil {
			if r.logger.Err() {
				r.logger.Err("borked response:", err)
			}
			return
		}
		go func() {
			if err := r.resolver.Resolve(&response); err != nil {
				if r.logger.Err() {
					r.logger.Err("failed to resolve:", err)
				}
			}
		}()
	}))
	if err != nil {
		if r.logger.Err() {
			r.logger.Err("OnMessage:", err)
		}
		return nil, err
	}
	if r.logger.Info() {
		r.logger.Info("listening on", r.responseChannel)
	}

	return r, nil
}

func (r *requester) Issue(pipeline []string, obj interface{}) (*Future, error) {

	if r.logger.Info() {
		r.logger.Info("issuing", pipeline, obj)
	}

	transaction := newTransaction(r.responseChannel, obj, pipeline[1:])
	to := pipeline[0]
	bytes, err := r.codec.Marshal(transaction)
	if err != nil {
		return nil, err
	}
	f := newFuture(transaction.ID)
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
	response chan *Transaction
	cached   *Transaction
	fetched  chan Signal
}

// newFuture creates a new response future that
// is initialized appropriately for waiting on an incoming
// response.
func newFuture(id RequestID) *Future {
	return &Future{id: id, response: make(chan *Transaction), fetched: make(chan Signal)}
}

// Response uses a future mechanism to retrieve the response.
// Execution continues asynchronously until this method is called,
// at which point execution blocks until the Response object is
// available, or if the timeout is reached.
// If the Response times out, nil is returned.
func (r *Future) Response(timeout time.Duration) (*Transaction, error) {
	select {
	case <-r.fetched: // response already here
		return r.cached, nil
	case r.cached = <-r.response: // response arrived
		close(r.fetched)
		return r.cached, nil
	case <-time.After(timeout):
		// timed out
		return nil, ErrTimeout
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
func (c *reqResolver) Resolve(response *Transaction) error {
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
