package qp

import "log"

// RequestHandler represents types capable of handling Requests.
type RequestHandler interface {
	Handle(*Request)
}

// RequestHandlerFunc represents functions capable of handling
// Requests.
type RequestHandlerFunc func(*Request)

// Handle calls the RequestHandlerFunc in order to handle
// the specific Request.
func (f RequestHandlerFunc) Handle(r *Request) {
	f(r)
}

// Responder responds to requests.
type Responder struct {
	name       string
	instanceID string
	uniqueID   string
	codec      Codec
	transport  DirectTransport
}

// NewResponder makes a new object capable of responding to requests.
func NewResponder(name, instanceID string, codec Codec, transport DirectTransport) *Responder {
	r := &Responder{
		codec:     codec,
		transport: transport,
		uniqueID:  name + "." + instanceID,
	}
	return r
}

// Handle binds a RequestHandler to the specified channel.
func (r *Responder) Handle(channel string, handler RequestHandler) error {

	r.transport.OnMessage(channel, HandlerFunc(func(msg *Message) {

		var request Request
		if err := r.codec.Unmarshal(msg.Data, &request); err != nil {
			log.Println("TODO: Handle unmarshal error", err)
			return
		}

		handler.Handle(&request)

		// at this point, the caller has mutated the data.
		// forward this request object to the next endpoint
		to := ""
		if len(request.To) != 0 {
			// pop off the first to
			to = request.To[0]
			request.To = request.To[1:]
		} else {
			// send it from form whence it came
			to = request.From[0]
		}
		request.From = append(request.From, r.uniqueID)

		// encode the data
		data, err := r.codec.Marshal(request)
		if err != nil {
			log.Println("Error encoding data for pipeline:", err.Error())
			return
		}

		// send the data
		r.transport.Send(to, data)

	}))

	return nil
}

// HandleFunc binds the specified function to the specified channel.
func (r *Responder) HandleFunc(channel string, f RequestHandlerFunc) error {
	return r.Handle(channel, f)
}
