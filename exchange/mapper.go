package exchange

import "sync"

// RequestHandler defines the function signature for the callback
// that will be called when a request is received.
type RequestHandler func(channel string, request *Request)

// Mapper uses a map internally to implement
// the mapper interface
type Mapper struct {
	lock  sync.RWMutex
	items map[string][]RequestHandler
}

// MakeMapper initializes and returns a mapper instance
// as a mapper interface.
func MakeMapper() *Mapper {
	return &Mapper{items: map[string][]RequestHandler{}}
}

// Track begins tracking an id and its associated handler so it
// can be found later
func (m *Mapper) Track(id string, handler RequestHandler) {
	m.lock.Lock()
	m.items[id] = append(m.items[id], handler)
	m.lock.Unlock()

}

// Find locates the given id and returns the handlers associated with it
func (m *Mapper) Find(id string) []RequestHandler {
	var handlers []RequestHandler
	m.lock.RLock()
	handlers = m.items[id]
	m.lock.RUnlock()
	return handlers
}
