package qp

import "sync"

// mapper uses a map internally to implement
// the mapper interface
type mapper struct {
	lock  sync.RWMutex
	items map[string][]RequestHandler
}

// makeMapper initializes and returns a mapper instance
// as a mapper interface.
func makeMapper() *mapper {
	return &mapper{items: map[string][]RequestHandler{}}
}

func (m *mapper) track(id string, handler RequestHandler) {
	m.lock.Lock()
	m.items[id] = append(m.items[id], handler)
	m.lock.Unlock()

}
func (m *mapper) find(id string) []RequestHandler {
	var handlers []RequestHandler
	m.lock.RLock()
	handlers = m.items[id]
	m.lock.RUnlock()
	return handlers
}
