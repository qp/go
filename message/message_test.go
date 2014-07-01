package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageNew(t *testing.T) {

	a := assert.New(t)

	m := New("service", "test")
	if a.NotNil(m) {
		a.NotEmpty(m.ID)
		a.Equal(m.From[0], "service/"+m.ID)
	}
}
