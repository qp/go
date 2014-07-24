package qp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolver(t *testing.T) {
	c := newResolver()

	id := unique()
	r := newResponse("test", "object", id)
	rf := newResponseFuture(id)

	if assert.NotNil(t, c) {
		c.Track(rf)
		go c.Resolve(r)
		r2 := rf.Response()
		assert.Equal(t, r, r2)
	}

}

func TestMapper(t *testing.T) {

	m := newReqMapper()
	if assert.NotNil(t, m) {

		assert.NotNil(t, m.items)

		run := false
		h := func(channel string, Request *Request) {
			run = true
		}
		m.Track("test", h)
		handlers := m.Find("test")

		if assert.NotNil(t, handlers) {
			handlers[0]("test", nil)
			assert.True(t, run)
		}
	}

}

func TestMapperMultiple(t *testing.T) {

	m := newReqMapper()
	if assert.NotNil(t, m) {

		assert.NotNil(t, m.items)

		run := false
		run2 := false
		h := func(channel string, Request *Request) {
			run = true
		}
		h2 := func(channel string, Request *Request) {
			run2 = true
		}

		m.Track("test", h)
		m.Track("test", h2)
		handlers := m.Find("test")

		if assert.NotNil(t, handlers) {
			handlers[0]("test", nil)
			assert.True(t, run)
			handlers[1]("test", nil)
			assert.True(t, run2)
		}
	}

}
