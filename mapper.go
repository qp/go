package qp

type mapper interface {
	track(string, RequestHandler)
	find(string) RequestHandler
}
