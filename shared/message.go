package shared

// Message is the standard QP messaging object.
// It is used to facilitate all communication between
// QP nodes, as well as containing the metadata
// necessary to implement the pipeline functionality.
type Message struct {
	To   StringDES   // array of destination addresses
	From StringDES   // array of addresses encountered thus far
	ID   string      // a UUID identifying this message
	Data interface{} // arbitrary data payload
	Err  interface{} // arbitrary error payload. nil if no error
}
