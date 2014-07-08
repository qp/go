package qp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapper(t *testing.T) {

	m := makeMapper()
	if assert.NotNil(t, m) {

		assert.NotNil(t, m.items)

		run := false
		h := func(channel string, request *Request) {
			run = true
		}
		m.track("test", h)
		handlers := m.find("test")

		if assert.NotNil(t, handlers) {
			handlers[0]("test", nil)
			assert.True(t, run)
		}
	}

}

func TestMapperMultiple(t *testing.T) {

	m := makeMapper()
	if assert.NotNil(t, m) {

		assert.NotNil(t, m.items)

		run := false
		run2 := false
		h := func(channel string, request *Request) {
			run = true
		}
		h2 := func(channel string, request *Request) {
			run2 = true
		}

		m.track("test", h)
		m.track("test", h2)
		handlers := m.find("test")

		if assert.NotNil(t, handlers) {
			handlers[0]("test", nil)
			assert.True(t, run)
			handlers[1]("test", nil)
			assert.True(t, run2)
		}
	}

}
