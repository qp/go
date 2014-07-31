package qp_test

import (
	"github.com/qp/go"
)

func json(obj interface{}) []byte {
	b, _ := qp.JSON.Marshal(obj)
	return b
}

func unjson(b []byte) interface{} {
	var o interface{}
	qp.JSON.Unmarshal(b, &o)
	return o
}
