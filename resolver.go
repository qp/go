package qp

import "sync"

// resolver is responsible for tracking futures
// and resolving them when a response is received
type resolver struct {
	items map[string]*ResponseFuture
	lock  sync.Mutex
}

// makeResolver creates and initializes a
// resolver object
func makeResolver() *resolver {
	return &resolver{items: map[string]*ResponseFuture{}}
}

// track begins tracking a ResponseFuture, waiting for
// a response to come in
func (c *resolver) track(future *ResponseFuture) {
	c.lock.Lock()
	c.items[future.id] = future
	c.lock.Unlock()
}

// resolve resolves a ResponseFuture by matching it up
// with the given Response
func (c *resolver) resolve(response *Response) {
	var future *ResponseFuture
	c.lock.Lock()
	future = c.items[response.ID]
	delete(c.items, response.ID)
	c.lock.Unlock()
	future.response <- response
}
