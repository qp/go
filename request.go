package qp

import (
	"strconv"
	"sync/atomic"
)

// Request defines all the fields and information
// in the standard qp request object. It is used
// as part of the RequestHandler callback.
type Request struct {
	*Response
	To []string `json:"to"` // array of destination addresses
}

var unique uint64

// newID returns a string representation of an atomically incrementing
// uint64.
func newID() string {
	id := atomic.AddUint64(&unique, 1)
	return strconv.FormatUint(id, 10)
}

// MakeRequest makes a new request object and generates a unique ID in the from array.
func MakeRequest(endpoint string, object interface{}, pipeline ...string) *Request {
	return &Request{To: pipeline, Response: makeResponse(endpoint, object, newID())}
}
