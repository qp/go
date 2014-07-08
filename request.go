package qp

import "github.com/qp/go/utils"

// Request defines all the fields and information
// in the standard qp request object. It is used
// as part of the RequestHandler callback.
type Request struct {
	*Response
	To []string `json:"to"` // array of destination addresses
}

// MakeRequest makes a new request object and generates a unique ID in the from array.
func MakeRequest(endpoint string, object interface{}, pipeline ...string) *Request {
	return &Request{To: pipeline, Response: makeResponse(endpoint, object, utils.UniqueStringID())}
}
