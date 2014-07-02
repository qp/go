package utils

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestDES(t *testing.T) {
	s := StringDES([]string{})
	s.Push("one")
	s.Push("two")
	s.Push("three")

	a := assert.New(t)
	a.Equal([]string{"three", "two", "one"}, []string(s))

	a.Equal("three", s.Peek())

	a.Equal("three", s.Pop())
	a.Equal(len(s), 2)
	a.Equal(cap(s), 2)
	a.Equal("two", s.Pop())
	a.Equal(len(s), 1)
	a.Equal(cap(s), 1)
	a.Equal("one", s.Pop())
	a.Equal(len(s), 0)
	a.Equal(cap(s), 0)

	s.BPush("three")
	s.BPush("two")
	s.BPush("one")
	a.Equal([]string{"three", "two", "one"}, []string(s))

	a.Equal("one", s.BPeek())

	a.Equal("one", s.BPop())
	a.Equal(len(s), 2)
	a.Equal(cap(s), 2)
	a.Equal("two", s.BPop())
	a.Equal(len(s), 1)
	a.Equal(cap(s), 1)
	a.Equal("three", s.BPop())
	a.Equal(len(s), 0)
	a.Equal(cap(s), 0)

}
