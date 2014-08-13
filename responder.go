package qp

import (
	"errors"

	"github.com/stretchr/slog"
)

// TransactionHandler represents types capable of handling Requests.
type TransactionHandler interface {
	Handle(req *Transaction) *Transaction
}

// TransactionHandlerFunc represents functions capable of handling
// Requests.
type TransactionHandlerFunc func(r *Transaction) *Transaction

// Handle calls the TransactionHandlerFunc in order to handle
// the specific Transaction.
func (f TransactionHandlerFunc) Handle(r *Transaction) *Transaction {
	return f(r)
}

// Responder represents types capable of responding to requests.
type Responder interface {
	// Handle binds a TransactionHandler to the specified channel.
	Handle(channel string, handler TransactionHandler) error
	// HandleFunc binds the specified function to the specified channel.
	HandleFunc(channel string, f TransactionHandlerFunc) error
}

// responder responds to requests.
type responder struct {
	name       string
	instanceID string
	uniqueID   string
	codec      Codec
	transport  DirectTransport
	log        slog.Logger
}

// NewResponder makes a new object capable of responding to requests.
func NewResponder(name, instanceID string, codec Codec, transport DirectTransport) Responder {
	return NewResponderLogger(name, instanceID, codec, transport, slog.NilLogger)
}

// NewResponderLogger makes a new object capable of responding to requests, which
// will log errors to the specified Logger.
func NewResponderLogger(name, instanceID string, codec Codec, transport DirectTransport, logger slog.Logger) Responder {
	return &responder{
		codec:     codec,
		transport: transport,
		uniqueID:  name + "." + instanceID,
		log:       logger,
	}
}

func (r *responder) Handle(channel string, handler TransactionHandler) error {

	return r.transport.OnMessage(channel, HandlerFunc(func(msg *Message) {

		var request Transaction
		if err := r.codec.Unmarshal(msg.Data, &request); err != nil {
			if r.log.Err() {
				r.log.Err("unmarshal error:", err)
			}
			return
		}

		request = *handler.Handle(&request)

		// at this point, the caller has mutated the data.
		// forward this request object to the next endpoint
		var to string
		if len(request.To) != 0 {
			// pop off the first to
			to = request.To[0]
			request.To = request.To[1:]
		} else {
			// send it from form whence it came
			if len(request.From) == 0 {
				err := errors.New("cannot respond when From field is empty")
				if r.log.Err() {
					r.log.Err("error handling request:", err)
				}
				return
			}
			to = request.From[0]
		}
		request.From = append(request.From, r.uniqueID)

		// encode the data
		data, err := r.codec.Marshal(request)
		if err != nil {
			if r.log.Err() {
				r.log.Err("error encoding data for pipeline:", err)
			}
			return
		}

		// send the data
		r.transport.Send(to, data)

	}))

}

func (r *responder) HandleFunc(channel string, f TransactionHandlerFunc) error {
	return r.Handle(channel, f)
}
