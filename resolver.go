package qp

// Resolver is an interface describing how
// futures are tracked and resolved
type resolver interface {
	track(*ResponseFuture)
	resolve(*Response)
}
