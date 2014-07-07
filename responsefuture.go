package qp

// ResponseFuture implements a future for a response object
// It allows execution to continue until the response object
// is requested from this object, at which point it blocks and
// waits for the response to come back.
type ResponseFuture struct {
	id       string
	response chan *Response
}

// MakeResponseFuture creates a new response future that
// is initialized appropriately for waiting on an incoming
// response.
func MakeResponseFuture(id string) *ResponseFuture {
	return &ResponseFuture{id: id, response: make(chan *Response)}
}
