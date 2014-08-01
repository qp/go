package bson

import (
	"github.com/qp/go"
	"labix.org/v2/mgo/bson"
)

// Codec is a Codec that talks BSON.
var Codec = qp.NewCodec(func(object interface{}) ([]byte, error) {
	return bson.Marshal(object)
}, func(data []byte, to interface{}) error {
	return bson.Unmarshal(data, to)
})
