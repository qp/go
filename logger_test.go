package qp_test

import (
	"fmt"
	"testing"

	"github.com/qp/go"
	"github.com/stretchr/testify/require"
)

type logger struct {
	output []string
}

func (l *logger) Error(v ...interface{}) {
	l.output = append(l.output, fmt.Sprint(v...))
}

var _ qp.Logger = (*logger)(nil)

func TestLoggers(t *testing.T) {

	l1 := &logger{}
	l2 := &logger{}
	l3 := &logger{}
	var ls qp.Logger
	ls = qp.Loggers(l1, l2, l3, qp.NilLogger)

	ls.Error("one")
	ls.Error("t", "w", "o")
	ls.Error("three")

	require.Equal(t, 3, len(l1.output))
	require.Equal(t, 3, len(l2.output))
	require.Equal(t, 3, len(l3.output))

	require.Equal(t, "one", l1.output[0])
	require.Equal(t, "two", l1.output[1])
	require.Equal(t, "three", l1.output[2])

	require.Equal(t, "one", l2.output[0])
	require.Equal(t, "two", l2.output[1])
	require.Equal(t, "three", l2.output[2])

	require.Equal(t, "one", l3.output[0])
	require.Equal(t, "two", l3.output[1])
	require.Equal(t, "three", l3.output[2])

}
