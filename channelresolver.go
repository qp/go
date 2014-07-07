package qp

// channelResolver is a kind of resolver that uses
// a go routine, channels and select statements
// to track and handle resolution of futures
type channelResolver struct {
}

// MakeChannelResolver creates and initializes a
// channelResolver object
func MakeChannelResolver() resolver {
	return &channelResolver{}
}

// Track begins tracking a ResponseFuture, waiting for
// a response to come in
func (c *channelResolver) track(*ResponseFuture) {

}

// Resolve resolves a ResponseFuture by matching it up
// with the given Response
func (c *channelResolver) resolve(*Response) {

}
