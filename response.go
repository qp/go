package qp

import (
	"time"

	"github.com/qp/go/messages"
)

// Response implements a simple future with a timeout.
// It is used to fetch the response from a pipeline request.
type Response struct {
	response chan *messages.Message
	timeout  time.Duration
}

//newResponse makes a new response object, creating
// the channel through which the response will be sent.
func newResponse(timeout time.Duration) *Response {
	return &Response{response: make(chan *messages.Message), timeout: timeout}
}

// Message blocks until a response is received or the timeout
// expires. If the timeout expires, the returned Message object
// will be empty except for the Err field being set to an error
// object.
func (r *Response) Message() *messages.Message {
	select {
	case m := <-r.response:
		return m
	case <-time.After(r.timeout):
		// TODO: fire a message to the messenger telling it to delete this from the map
		return &messages.Message{Err: map[string]interface{}{"message": "timeout expired while waiting for response"}}
	}
}
