package qp_test

import (
	"testing"

	"github.com/qp/go"
	"github.com/stretchr/testify/assert"
)

func TestJson(t *testing.T) {

	var c qp.Codec
	c = qp.JSON
	obj := map[string]interface{}{"key": "value"}

	m, err := c.Marshal(obj)
	if assert.NoError(t, err) {
		if assert.Equal(t, `{"key":"value"}`, string(m)) {
			var um map[string]interface{}
			if err := c.Unmarshal(m, &um); assert.NoError(t, err) {
				assert.Equal(t, obj, um)
			}
		}
	}

}
