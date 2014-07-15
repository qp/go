package exchange

import "sync"

// RequestResolver is responsible for tracking futures
// and resolving them when a response is received
type RequestResolver struct {
	items map[string]*ResponseFuture
	lock  sync.Mutex
}

// MakeResolver creates and initializes a
// resolver object
func MakeResolver() *RequestResolver {
	return &RequestResolver{items: map[string]*ResponseFuture{}}
}

// Track begins tracking a ResponseFuture, waiting for
// a response to come in
func (c *RequestResolver) Track(future *ResponseFuture) {
	c.lock.Lock()
	c.items[future.id] = future
	c.lock.Unlock()
}

// Resolve resolves a ResponseFuture by matching it up
// with the given Response
func (c *RequestResolver) Resolve(response *Response) {
	var future *ResponseFuture
	c.lock.Lock()
	future = c.items[response.ID]
	delete(c.items, response.ID)
	c.lock.Unlock()
	future.response <- response
}
