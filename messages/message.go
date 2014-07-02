package messages

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/qp/go/utils"
)

// Message is the standard QP messaging object.
// It is used to facilitate all communication between
// QP nodes, as well as containing the metadata
// necessary to implement the pipeline functionality.
type Message struct {
	To   utils.StringDES `json:"to"`            // array of destination addresses
	From utils.StringDES `json:"from"`          // array of addresses encountered thus far
	ID   string          `json:"id"`            // a UUID identifying this message
	Data interface{}     `json:"data"`          // arbitrary data payload
	Err  interface{}     `json:"err,omitempty"` // arbitrary error payload. nil if no error
}

// NewMessage creates a new Message object with appropriate fields set.
func NewMessage(serviceName string, data interface{}, to ...string) *Message {
	id := uuid.New()
	return &Message{To: to, From: []string{serviceName + "/" + id}, ID: id, Data: data}
}
