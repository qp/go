package event

import (
	"log"

	"github.com/qp/go/transports"
)

// Log implements a transport that simply logs
// activity to console.
type Log struct {
}

// MakeLog makes and initializes a new log transport
func MakeLog() transports.EventTransport {
	return &Log{}
}

// Publish logs activity
func (l *Log) Publish(to string, data []byte) error {
	log.Println("Publishing", string(data), "to:", to)
	return nil
}

// ListenFor logs activity
func (l *Log) ListenFor(channel string) {
	log.Println("Listening on channel:", channel)
}

// OnMessage logs activity
func (l *Log) OnMessage(messageFunc transports.MessageFunc) {
	log.Println("OnMessage")
}

// Start logs activity
func (l *Log) Start() error {
	log.Println("Start")
	return nil
}

// Stop logs activity
func (l *Log) Stop() {
	log.Println("Stop")
}
