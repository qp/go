package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageNew(t *testing.T) {

	a := assert.New(t)
	d := map[string]interface{}{"name": "Tyler"}
	m := New("service", d, "test")
	if a.NotNil(m) {
		a.NotEmpty(m.ID)
		a.Equal(m.From[0], "service/"+m.ID)
		a.Equal(d, m.Data)
	}
}
