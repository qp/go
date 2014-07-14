package request

import (
	"log"

	"github.com/qp/go/transports"
)

// Log implements a transport that simply logs
// activity to console.
type Log struct {
	quiet bool
}

// MakeLog makes and initializes a new log transport
func MakeLog(quiet bool) transports.RequestTransport {
	return &Log{quiet: quiet}
}

func (l *Log) shouldLog(args ...interface{}) {
	if !l.quiet {
		log.Println(args...)
	}
}

// Send logs activity
func (l *Log) Send(to string, data []byte) error {
	l.shouldLog("Sending", string(data), "to:", to)
	return nil
}

// ListenFor logs activity
func (l *Log) ListenFor(channel string) {
	l.shouldLog("Listening on channel:", channel)
}

// OnMessage logs activity
func (l *Log) OnMessage(messageFunc transports.MessageFunc) {
	l.shouldLog("OnMessage")
}

// Start logs activity
func (l *Log) Start() error {
	l.shouldLog("Start")
	return nil
}

// Stop logs activity
func (l *Log) Stop() {
	l.shouldLog("Stop")
}
