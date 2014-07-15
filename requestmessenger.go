package qp

import (
	"log"

	"code.google.com/p/go-uuid/uuid"
	"github.com/qp/go/codecs"
	"github.com/qp/go/exchange"
	"github.com/qp/go/transports"
)

// RequestMessenger defines the interface through which
// requests are introduced into the qp system, and responses
// to those requests are returned.
// RequestMessenger implements the RequestMessenger interface
type RequestMessenger struct {
	name         string
	responseName string
	codec        codecs.Codec
	transport    transports.RequestTransport
	resolver     *exchange.RequestResolver
	mapper       *exchange.RequestMapper
}

// MakeRequestMessenger creates a new request messenger that allows direct communication between
// two endpoints in the qp system. It also allows for pipelining through multiple specified
// endpoints.
func MakeRequestMessenger(name, instanceName string, codec codecs.Codec, transport transports.RequestTransport) *RequestMessenger {
	if instanceName == "" {
		instanceName = uuid.New()
	}

	r := &RequestMessenger{name: name,
		responseName: name + "." + instanceName,
		codec:        codec,
		transport:    transport,
		resolver:     exchange.MakeResolver(),
		mapper:       exchange.MakeRequestMapper(),
	}

	// listen on the "responseName" channel
	r.transport.ListenFor(r.responseName)

	r.transport.OnMessage(func(bm *transports.BinaryMessage) {
		// switch on the bm.channel to determine the type of message
		if bm.Channel == r.responseName {
			// decode to response object
			var response exchange.Response
			err := r.codec.Unmarshal(bm.Data, &response)
			if err != nil {
				// dispatch a log entry and abort
				log.Println("Unable to unmarshal response: ", err)
				return
			}

			// map the response to the appropriate ResponseFuture
			go r.resolver.Resolve(&response)
		} else {
			// decode to request object
			var request exchange.Request
			err := r.codec.Unmarshal(bm.Data, &request)
			if err != nil {
				log.Println("Unable to unmarshal request: ", err)
				return
			}

			// map the request to the appropriate RequestHandler
			handlers := r.mapper.Find(bm.Channel)
			if handlers != nil {
				go func() {
					for _, handler := range handlers {
						// call each RequestHandler
						handler(bm.Channel, &request)
					}

					// get the next destination endpoint
					to := ""
					if len(request.To) != 0 {
						to = request.To[0]
						request.To = request.To[1:]
					} else {
						to = request.From[0]
					}
					request.From = append(request.From, r.responseName)

					// encode the request
					data, err := r.codec.Marshal(&request)
					if err != nil {
						log.Println("Unable to marshal request to send to next endpoint: ", err)
						return
					}

					// send the request to the next endpoint in the pipeline
					err = r.transport.Send(to, data)
					if err != nil {
						log.Println("Unable to send request to next endpoint: ", err)
						return
					}
				}()
			}
		}
	})

	return r
}

// Start spins up the Request Messenger to start processing
// incoming and outgoing messages.
func (r *RequestMessenger) Start() {
	// spin up the underlying transport
	r.transport.Start()
}

// Stop tears down the Request Messenger in a graceful manner, allowing
// any in-flight operations to complete.
// After calling Stop, you may call Start again to resume processing. However,
// typically this method is called only once.
func (r *RequestMessenger) Stop() {
	// spin down the underlying transport
	r.transport.Stop()
}

// OnRequest assigns the given handler to the given channels, calling the handler
// when a message is received on one of those channels.
func (r *RequestMessenger) OnRequest(handler exchange.RequestHandler, channels ...string) {

	// validate handler is not nil
	if handler == nil {
		panic("handler cannot be nil")
	}
	// validate channels is not empty
	if len(channels) == 0 {
		panic("channels cannot be empty")
	}

	// associate each channel with the appropriate handler function
	for _, channel := range channels {
		r.mapper.Track(channel, handler)
		// instruct the transport to listen on the channel
		r.transport.ListenFor(channel)
	}

}

// Request constructs a request to be sent to the given pipeline endpoint(s). The pipeline may
// be one or more endpoints. If it is more than one, each will receive the message, in order, and
// have an opportunity to mutate it before it is dispatched to the next endpoint in the pipeline.
// The provided object will be serialized and send as the "data" field in the message.
func (r *RequestMessenger) Request(object interface{}, pipeline ...string) (*exchange.ResponseFuture, error) {

	// validate that we have a pipeline
	if len(pipeline) == 0 {
		panic("pipeline cannot be empty")
	}
	//validate that we have an object
	if object == nil {
		panic("object cannot be nil")
	}

	// create a new request message object
	// assign the given "object" to the "data" field in the request object
	// assign the "to" stack in the request object using the pipeline string, except the
	// top that has been poppped off and is being used to make the transport call
	request := exchange.MakeRequest(r.responseName, object, pipeline[1:]...)
	to := pipeline[0]

	// encode the request object to a byte slice
	data, err := r.codec.Marshal(request)
	if err != nil {
		return nil, err
	}

	// use the unique ID in the request object to associate a request with a reply
	// we have to map the request to the response future, then handle that response when it comes back
	rf := exchange.MakeResponseFuture(request.ID)
	r.resolver.Track(rf)

	// call the transport and give it the popped "to" endpoint, as well as
	// the request object to that endpoint and give it the encoded message
	err = r.transport.Send(to, data)
	if err != nil {
		return nil, err
	}

	return rf, nil
}
