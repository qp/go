package utils

import (
	"strconv"
	"sync/atomic"
)

var unique uint64

// UniqueID returns a unique uint64.
func UniqueID() uint64 {
	return atomic.AddUint64(&unique, 1)
}

// UniqueStringID returns a unique uint64 as a string.
func UniqueStringID() string {
	return strconv.FormatUint(UniqueID(), 10)
}
