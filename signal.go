package qp

// Signal is an empty struct type useful for
// signalling down channels.
type Signal struct{}

// SignalNow is an instance of Signal.
var SignalNow = Signal{}
