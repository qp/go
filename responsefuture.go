package qp

// ResponseFuture implements a future for a response object
// It allows execution to continue until the response object
// is requested from this object, at which point it blocks and
// waits for the response to come back.
type ResponseFuture struct {
	id       string
	response chan *Response
}

// makeResponseFuture creates a new response future that
// is initialized appropriately for waiting on an incoming
// response.
func makeResponseFuture(id string) *ResponseFuture {
	return &ResponseFuture{id: id, response: make(chan *Response)}
}

// Response uses a future mechanism to retrieve the response.
// Execution continues asynchronously until this method is called,
// at which point execution blocks until the Response object is
// available.
//
// There is no timeout. It will block indefinitely. This may
// change in the future.
func (r *ResponseFuture) Response() *Response {
	return <-r.response
}
