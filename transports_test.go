package qp_test

import (
	"testing"

	"github.com/qp/go"
)

func TestHandlerFunc(t *testing.T) {

	var _ qp.Handler = HandlerFunc(func(m *qp.Message) {})

}
