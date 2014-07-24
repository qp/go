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

// ensure jsonCodec is a Codec
var _ Codec = (*jsonCodec)(nil)

// JSON is a Codec that talks JSON.
var JSON *jsonCodec

// jsonCodec is a Codec that talks JSON.
type jsonCodec struct{}

// Marshal an object into a JSON byte slice representation
func (_ *jsonCodec) Marshal(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

// Unmarshal an object from a JSON byte slice into an object
func (_ *jsonCodec) Unmarshal(data []byte, to interface{}) error {
	return json.Unmarshal(data, to)
}
