package exchange

import "github.com/qp/go/utils"

// Request defines all the fields and information
// in the standard qp request object. It is used
// as part of the RequestHandler callback.
type Request struct {
	To   []string    `json:"to"`   // array of destination addresses
	From []string    `json:"from"` // array of addresses encountered thus far
	ID   string      `json:"id"`   // a UUID identifying this message
	Data interface{} `json:"data"` // arbitrary data payload
}

// MakeRequest makes a new request object and generates a unique ID in the from array.
func MakeRequest(endpoint string, object interface{}, pipeline ...string) *Request {
	return &Request{To: pipeline, From: []string{endpoint}, ID: utils.UniqueStringID(), Data: object}
}
