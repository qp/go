package qp

import "github.com/stretchr/pat/stop"

// StartStopper represents an object that can be started and
// stopped gracefully.
type StartStopper interface {
	Start() error
	stop.Stopper
}
