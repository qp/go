package transports

import (
	"github.com/qp/go/transports/common"
	"github.com/qp/go/transports/request"
)

// Kind is a type used to define what kind of
// transport to create in the Make* functions.
type Kind uint8

const (
	// KindRequest specifies a "Request" transport
	KindRequest Kind = iota
	// KindEvent pecifies an "Event" transport
	KindEvent
)

// MakeInProc creates a new instance of an InProc transport
// of the given Kind
func MakeInProc(kind Kind, wrapped common.Transport) common.Transport {
	switch kind {
	case KindRequest:
		return request.MakeInProc(wrapped)
	case KindEvent:
	}
	return nil
}

// MakeLog makes and initializes a new log transport of the given kind
func MakeLog(kind Kind, quiet bool) common.Transport {
	switch kind {
	case KindRequest:
		return request.MakeLog(quiet)
	case KindEvent:
	}
	return nil
}

// MakeRedis initializes a new Redis transport instance of the given kind
func MakeRedis(kind Kind, url string) common.Transport {
	switch kind {
	case KindRequest:
		return request.MakeRedis(url)
	case KindEvent:
	}
	return nil
}
