package exchange

import "fmt"

// Event defines all the fields and information
// included as part of a Event to a request.
type Event struct {
	From string      `json:"from"` // address of sender
	Data interface{} `json:"data"` // arbitrary data payload
}

// MakeEvent makes a new Event object
func MakeEvent(endpoint string, object interface{}) *Event {
	return &Event{From: endpoint, Data: object}
}

func (r Event) String() string {
	return fmt.Sprintf("From: %v\nData: %v", r.From, r.Data)
}
