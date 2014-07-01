package codecs

import (
	"encoding/json"
)

// ensure JSON conforms to Codec
var conformJSON = Codec(&JSON{})

// JSON is the JSON implementation of the Codec interface
type JSON struct{}

// Marshal an object into a JSON byte slice representation
func (j *JSON) Marshal(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

// Unmarshal an object from a JSON byte slice into an object
func (j *JSON) Unmarshal(data []byte, to interface{}) error {
	return json.Unmarshal(data, to)
}