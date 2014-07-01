package shared

import (
	"code.google.com/p/go-uuid/uuid"

	"github.com/qp/go/shared"
)

// Message is the standard QP messaging object.
// It is used to facilitate all communication between
// QP nodes, as well as containing the metadata
// necessary to implement the pipeline functionality.
type Message struct {
	To   shared.StringDES // array of destination addresses
	From shared.StringDES // array of addresses encountered thus far
	ID   string           // a UUID identifying this message
	Data interface{}      // arbitrary data payload
	Err  interface{}      // arbitrary error payload. nil if no error
}

// New creates a new Message object with appropriate fields set.
func New(serviceName string, data interface{}, to ...string) *Message {
	id := uuid.New()
	return &Message{To: to, From: []string{serviceName + "/" + id}, ID: id, Data: data}
}
