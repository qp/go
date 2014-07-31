go [![GoDoc](https://godoc.org/github.com/qp/go?status.png)](http://godoc.org/github.com/qp/go) [![wercker status](https://app.wercker.com/status/45dc654cf4ed2c704e5beaa45a1c716d/s "wercker status")](https://app.wercker.com/project/bykey/45dc654cf4ed2c704e5beaa45a1c716d)
==

The Go implementation of the QP protocol.

### Usage

See the [example code](https://github.com/qp/go/tree/master/example) for a working example of how to implement Go services using QP.

#### Publish

Use a `NewPublisher` to publish events.

```go
// make a transport
transport := redis.NewPubSub("127.0.0.1:6379")
transport.Start()
defer transport.Stop()

// make a publisher
pub := qp.NewPublisher("name", "instanceID", qp.JSON, transport)

// publish messages
pub.Publish("channel1", "Hello QP")
pub.Publish("channel2", "Bonjour QP")
pub.Publish("channel3", "¡hola! QP")
```

#### Subscribe

Use a `NewSubscriber` to subscribe to events.

```go
// make a transport
transport := redis.NewPubSub("127.0.0.1:6379")
transport.Start()
defer transport.Stop()

// make a subscriber
sub := qp.NewSubscriber(qp.JSON, transport)

// subscribe to messages
sub.SubscribeFunc("channel1", func(event *qp.Event) {
  log.Println(event)
})
sub.SubscribeFunc("channel2", func(event *qp.Event) {
  log.Println(event)
})
sub.SubscribeFunc("channel3", func(event *qp.Event) {
  log.Println(event)
})
```

#### Request

Use a `NewRequester` to make requests.

```go
// make a transport
transport := redis.NewDirect("127.0.0.1:6379")
transport.Start()
defer transport.Stop()

// make a requester
req := qp.NewRequester("webserver", "one", qp.JSON, t)

// issue a request and wait 1 second for the response
response := req.Issue([]string{"channel1","channel2","channel3"}, "some data").Response(1 * time.Second)

log.Println(response)
```

#### Responders

Use a `NewResponder` to respond to requests.

```go
// make a transport
transport := redis.NewDirect("127.0.0.1:6379")
transport.Start()
defer transport.Stop()

res := qp.NewResponder("service", "one", qp.JSON, t)
res.HandleFunc("channel1", func(r *qp.Request) {
  // do some work and update r.Data
})
res.HandleFunc("channel2", func(r *qp.Request) {
  // do some work and update r.Data
})
```

#### Service

A `Service` is a special `Responder` that responds to requests on a channel
of its own name.

```go
// make a transport
transport := redis.NewDirect("127.0.0.1:6379")
transport.Start()
defer transport.Stop()

qp.ServiceFunc("serviceName", "instance", qp.JSON, transport, func(r *qp.Request) {
  // provide your service
})
```