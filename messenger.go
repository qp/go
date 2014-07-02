package qp

import "github.com/qp/go/codecs"

type Messenger struct {
	codec     codecs.Codec
	transport transports.Transport
}
