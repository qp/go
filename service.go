package qp

// Service implements the concept of a simple "Service".
//
// A Service is an endpoint that automatically subscribes
// to its own name, allowing other endpoints to issue
// requests to it. Multiple services with the same name
// will automatically draw upon the same channel, creating
// implicit load balancing.
//
// A Service may only use a Direct transport.
//
// By default, a Service will simply forward the message to
// the next endpoint.
type Service struct {
	Handler   RequestHandler
	responder Responder
}

// NewService creates a new Service. It automatically subscribes to a channel
// of its own name and forwards messages by default. If a "Handler" is set, it
// will be called.
func NewService(name, instanceID string, codec Codec, transport DirectTransport) *Service {
	s := &Service{
		responder: NewResponder(name, instanceID, codec, transport),
	}
	s.responder.HandleFunc(name, func(r *Request) {
		if s.Handler != nil {
			s.Handler.Handle(r)
		}
	})
	return s
}
