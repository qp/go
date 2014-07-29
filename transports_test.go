package qp_test

import (
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

// TestPubSubTransport is a mock qp.PubSubTransport
type TestPubSubTransport struct {
	Published  map[string][]byte
	Subscribed map[string]qp.Handler
	Err        error
	Running    bool
}

var _ qp.PubSubTransport = (*TestPubSubTransport)(nil)

func (t *TestPubSubTransport) Publish(c string, d []byte) error {
	if t.Published == nil {
		t.Published = make(map[string][]byte)
	}
	t.Published[c] = d
	return t.Err
}
func (t *TestPubSubTransport) Subscribe(c string, h qp.Handler) error {
	if t.Subscribed == nil {
		t.Subscribed = make(map[string]qp.Handler)
	}
	t.Subscribed[c] = h
	return t.Err
}
func (t *TestPubSubTransport) Start() error {
	t.Running = true
	return nil
}
func (t *TestPubSubTransport) Stop(wait time.Duration) {
	t.Running = false
}
func (t *TestPubSubTransport) StopChan() <-chan stop.Signal { return stop.Stopped() }

type TestDirectTransport struct {
	Sends      map[string][]byte
	OnMessages map[string]qp.Handler
	Err        error
	Running    bool
}

var _ qp.DirectTransport = (*TestDirectTransport)(nil)

func (t *TestDirectTransport) OnMessage(s string, h qp.Handler) error {
	if t.OnMessages == nil {
		t.OnMessages = make(map[string]qp.Handler)
	}
	t.OnMessages[s] = h
	return t.Err
}
func (t *TestDirectTransport) Send(s string, d []byte) error {
	if t.Sends == nil {
		t.Sends = make(map[string][]byte)
	}
	t.Sends[s] = d
	return t.Err
}
func (t *TestDirectTransport) Start() error {
	t.Running = true
	return nil
}
func (t *TestDirectTransport) Stop(time.Duration) {
	t.Running = false
}
func (t *TestDirectTransport) StopChan() <-chan stop.Signal { return stop.Stopped() }

func TestRequestHandlerFunc(t *testing.T) {

	var _ qp.Handler = qp.HandlerFunc(func(m *qp.Message) {})

}
