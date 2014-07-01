package codecs

// Codec is an interface declaring functions used for
// encoding and decoding an object to and from a given
// format, such as JSON
type Codec interface {
	// Marshal takes an object and creates a byte slice representation
	// of the object in the underlying data format.
	Marshal(object interface{}) ([]byte, error)
	// Unmarshal takes a bytes slice of data in the underlying data format
	// and decodes it into the provided object
	Unmarshal(data []byte, to interface{}) error
}
