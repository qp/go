// Package qp implements the "QP" protocol found at qp.github.io
//
// The QP protocol allows for agnostic communication with any underlying
// message queue system. By using QP, you remove technology-dependent
// code from your application, while gaining additional functionality,
// such as QP's pipeline concept.
//
// PubSub
//
// The publish/subscribe model is achieved by using the Publisher and
// Subscriber types, which expose Publish and Subscribe methods
// respectively.
//
// Request and Response
//
// Making requests and getting back a response from a pipeline of handlers
// is handled by using the Requester type, which offers the Issue
// method.
//
// Building services that respond to requests can be achieved by using
// the Responder type, which exposes the Handle method.
package qp
