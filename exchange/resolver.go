package exchange

import "sync"

// Resolver is responsible for tracking futures
// and resolving them when a response is received
type Resolver struct {
	items map[string]*ResponseFuture
	lock  sync.Mutex
}

// MakeResolver creates and initializes a
// resolver object
func MakeResolver() *Resolver {
	return &Resolver{items: map[string]*ResponseFuture{}}
}

// Track begins tracking a ResponseFuture, waiting for
// a response to come in
func (c *Resolver) Track(future *ResponseFuture) {
	c.lock.Lock()
	c.items[future.id] = future
	c.lock.Unlock()
}

// Resolve resolves a ResponseFuture by matching it up
// with the given Response
func (c *Resolver) Resolve(response *Response) {
	var future *ResponseFuture
	c.lock.Lock()
	future = c.items[response.ID]
	delete(c.items, response.ID)
	c.lock.Unlock()
	future.response <- response
}
