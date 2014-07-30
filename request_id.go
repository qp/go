package qp

import "sync/atomic"

// RequestID represents a unique ID for a Request.
type RequestID uint64

// u holds the last unique number.
var u uint64

// unique returns a unique uint64.
func unique() RequestID {
	return RequestID(atomic.AddUint64(&u, 1))
}
