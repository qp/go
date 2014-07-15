package exchange

import "fmt"

// Response defines all the fields and information
// included as part of a response to a request.
type Response struct {
	From []string    `json:"from"` // array of addresses encountered thus far
	ID   string      `json:"id"`   // a UUID identifying this message
	Data interface{} `json:"data"` // arbitrary data payload
}

// MakeResponse makes a new response object and generates a unique ID in the
// from array.
func MakeResponse(endpoint string, object interface{}, id string) *Response {
	return &Response{From: []string{endpoint}, ID: id, Data: object}
}

func (r Response) String() string {
	return fmt.Sprintf("From: %v\nID: %s\nData: %v", r.From, r.ID, r.Data)
}
