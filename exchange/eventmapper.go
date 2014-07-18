package exchange

import (
	"strings"
	"sync"
)

// EventHandler defines the function signature for the callback
// that will be called when a event is received.
type EventHandler func(channel string, event *Event)

// EventMapper uses a map internally to implement
// the mapper interface
type EventMapper struct {
	lock  sync.RWMutex
	items map[string][]EventHandler
}

// MakeEventMapper initializes and returns a mapper instance
// as a mapper interface.
func MakeEventMapper() *EventMapper {
	return &EventMapper{items: map[string][]EventHandler{}}
}

// Track begins tracking an id and its associated handler so it
// can be found later
func (m *EventMapper) Track(id string, handler EventHandler) {
	m.lock.Lock()
	m.items[id] = append(m.items[id], handler)
	m.lock.Unlock()

}

// Find locates the given id and returns the handlers associated with it
func (m *EventMapper) Find(id string) []EventHandler {
	var handlers []EventHandler
	m.lock.RLock()
	for itemID, item := range m.items {
		if strings.HasSuffix(itemID, "*") {
			if strings.HasPrefix(id, itemID[:len(itemID)-1]) {
				handlers = append(handlers, item...)
			}
		} else if id == itemID {
			handlers = append(handlers, item...)
		}
	}
	m.lock.RUnlock()
	return handlers
}
