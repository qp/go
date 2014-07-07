package qp

// RequestHandler defines the function signature for the callback
// that will be called when a request is received.
type RequestHandler func(channel string, request *Request)

// RequestMessenger defines the interface through which
// requests are introduced into the qp system, and responses
// to those requests are returned.
type RequestMessenger interface {
	Start()
	Stop()
	OnRequest(handler RequestHandler, channels ...string)
	Request(object interface{}, pipeline ...string) (*ResponseFuture, error)
}
