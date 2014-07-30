package qp

// RequestHandler represents types capable of handling Requests.
type RequestHandler interface {
	Handle(r *Request)
}

// RequestHandlerFunc represents functions capable of handling
// Requests.
type RequestHandlerFunc func(r *Request)

// Handle calls the RequestHandlerFunc in order to handle
// the specific Request.
func (f RequestHandlerFunc) Handle(r *Request) {
	f(r)
}

// Responder represents types capable of responding to requests.
type Responder interface {
	// Handle binds a RequestHandler to the specified channel.
	Handle(channel string, handler RequestHandler) error
	// HandleFunc binds the specified function to the specified channel.
	HandleFunc(channel string, f RequestHandlerFunc) error
}

// responder responds to requests.
type responder struct {
	name       string
	instanceID string
	uniqueID   string
	codec      Codec
	transport  DirectTransport
	log        Logger
}

// NewResponder makes a new object capable of responding to requests.
func NewResponder(name, instanceID string, codec Codec, transport DirectTransport) Responder {
	return NewResponderLogger(name, instanceID, codec, transport, NilLogger)
}

// NewResponderLogger makes a new object capable of responding to requests, which
// will log errors to the specified Logger.
func NewResponderLogger(name, instanceID string, codec Codec, transport DirectTransport, logger Logger) Responder {
	return &responder{
		codec:     codec,
		transport: transport,
		uniqueID:  name + "." + instanceID,
		log:       logger,
	}
}

func (r *responder) Handle(channel string, handler RequestHandler) error {

	r.transport.OnMessage(channel, HandlerFunc(func(msg *Message) {

		var request Request
		if err := r.codec.Unmarshal(msg.Data, &request); err != nil {
			r.log.Println("TODO: Handle unmarshal error", err)
			return
		}

		handler.Handle(&request)

		// at this point, the caller has mutated the data.
		// forward this request object to the next endpoint
		var to string
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
			r.log.Println("Error encoding data for pipeline:", err.Error())
			return
		}

		// send the data
		r.transport.Send(to, data)

	}))

	return nil
}

func (r *responder) HandleFunc(channel string, f RequestHandlerFunc) error {
	return r.Handle(channel, f)
}
