package qp

import "sync/atomic"

// u holds the last unique number.
var u uint64

// unique returns a unique uint64.
func unique() RequestID {
	return RequestID(atomic.AddUint64(&u, 1))
}
