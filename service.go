package qp

import "github.com/stretchr/slog"

// Service is an endpoint that automatically subscribes
// to its own name, allowing other endpoints to issue
// requests to it. Multiple services with the same name
// will automatically draw upon the same channel, creating
// implicit load balancing.
func Service(name, instanceID string, codec Codec, transport DirectTransport, handler TransactionHandler) error {
	return ServiceLogger(name, instanceID, codec, transport, slog.NilLogger, handler)
}

// ServiceFunc creates a service with a TransactionHandlerFunc rather than a
// TransactionHandler.
func ServiceFunc(name, instanceID string, codec Codec, transport DirectTransport, handler TransactionHandlerFunc) error {
	// TODO: test this
	return Service(name, instanceID, codec, transport, handler)
}

// ServiceLogger does the same thing as Service but also uses the
// specified Logger to log to.
func ServiceLogger(name, instanceID string, codec Codec, transport DirectTransport, logger slog.Logger, handler TransactionHandler) error {
	return NewResponderLogger(name, instanceID, codec, transport, logger).Handle(name, handler)
}

// ServiceLoggerFunc does the same thing ServiceLogger does but takes a
// TransactionHandlerFunc rather than a TransactionHandler.
func ServiceLoggerFunc(name, instanceID string, codec Codec, transport DirectTransport, logger slog.Logger, handler TransactionHandlerFunc) error {
	// TODO: test this
	return ServiceLogger(name, instanceID, codec, transport, logger, handler)
}
