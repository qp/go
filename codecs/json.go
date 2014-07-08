package codecs

import (
	"encoding/json"
)

// ensure JSON conforms to Codec
var _ Codec = (*JSON)(nil)

// JSON is the JSON implementation of the Codec interface
type JSON struct{}

// MakeJSON makes a new JSON codec and returns it
func MakeJSON() *JSON {
	return &JSON{}
}

// Marshal an object into a JSON byte slice representation
func (j *JSON) Marshal(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

// Unmarshal an object from a JSON byte slice into an object
func (j *JSON) Unmarshal(data []byte, to interface{}) error {
	return json.Unmarshal(data, to)
}
