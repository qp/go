package qp

// Service is an endpoint that automatically subscribes
// to its own name, allowing other endpoints to issue
// requests to it. Multiple services with the same name
// will automatically draw upon the same channel, creating
// implicit load balancing.
//
// By default, a Service will simply forward the message to
// the next endpoint, unless the request is mutated by the
// handler.
func Service(name, instanceID string, codec Codec, transport DirectTransport, handler RequestHandler) {
	ServiceLogger(name, instanceID, codec, transport, NilLogger, handler)
}

// ServiceLogger does the same thing as Service but also uses the
// specified Logger to log to.
func ServiceLogger(name, instanceID string, codec Codec, transport DirectTransport, logger Logger, handler RequestHandler) {
	NewResponderLogger(name, instanceID, codec, transport, logger).Handle(name, handler)
}
