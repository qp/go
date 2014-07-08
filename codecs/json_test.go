package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodecsJSON(t *testing.T) {

	c := Codec(&JSON{})
	a := assert.New(t)

	m := map[string]interface{}{"name": "Tyler"}
	bytes, err := c.Marshal(m)
	if a.NoError(err) {
		a.Equal(`{"name":"Tyler"}`, string(bytes))
	}

	var o map[string]interface{}
	err = c.Unmarshal(bytes, &o)
	if a.NoError(err) {
		a.Equal(m, o)
	}
}
