package qp

import (
	"encoding/json"
)

// Codec defines types that can marshal and unmarshal data to and from
// bytes.
type Codec interface {
	// Marshal takes an object and creates a byte slice representation
	// of the object in the underlying data format.
	Marshal(object interface{}) ([]byte, error)
	// Unmarshal takes a bytes slice of data in the underlying data format
	// and decodes it into the provided object
	Unmarshal(data []byte, to interface{}) error
}

type codec struct {
	marshal   func(object interface{}) ([]byte, error)
	unmarshal func(data []byte, to interface{}) error
}

func (c *codec) Marshal(object interface{}) ([]byte, error) {
	return c.marshal(object)
}
func (c *codec) Unmarshal(data []byte, to interface{}) error {
	return c.unmarshal(data, to)
}

// NewCodec makes a new Codec with the specified marshal and
// unmarshal functions.
func NewCodec(marshal func(object interface{}) ([]byte, error), unmarshal func(data []byte, to interface{}) error) Codec {
	return &codec{marshal: marshal, unmarshal: unmarshal}
}

// JSON is a Codec that talks JSON.
var JSON = NewCodec(func(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}, func(data []byte, to interface{}) error {
	return json.Unmarshal(data, to)
})
