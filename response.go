package qp

import "fmt"

// Response defines all the fields and information
// included as part of a response to a request.
type Response struct {
	// From is an array of addresses encountered thus far
	From []string `json:"from"`
	// ID is the ID of the request to which this response relates
	ID RequestID `json:"id"`
	// Data is the repsonse data payload
	Data interface{} `json:"data"`
}

// newResponse makes a new response object
func newResponse(endpoint string, object interface{}, id RequestID) *Response {
	return &Response{From: []string{endpoint}, ID: id, Data: object}
}

func (r Response) String() string {
	return fmt.Sprintf("From: %v\nID: %v\nData: %v", r.From, r.ID, r.Data)
}

// ResponseFuture implements a future for a response object
// It allows execution to continue until the response object
// is requested from this object, at which point it blocks and
// waits for the response to come back.
type ResponseFuture struct {
	id       RequestID
	response chan *Response
	cached   *Response
	fetched  chan struct{}
}

// newResponseFuture creates a new response future that
// is initialized appropriately for waiting on an incoming
// response.
func newResponseFuture(id RequestID) *ResponseFuture {
	return &ResponseFuture{id: id, response: make(chan *Response), fetched: make(chan struct{})}
}

// Response uses a future mechanism to retrieve the response.
// Execution continues asynchronously until this method is called,
// at which point execution blocks until the Response object is
// available.
//
// There is no timeout. It will block indefinitely. This may
// change in the future.
func (r *ResponseFuture) Response() *Response {
	select {
	case <-r.fetched:
		return r.cached
	case r.cached = <-r.response:
		close(r.fetched)
		return r.cached
	}
}
