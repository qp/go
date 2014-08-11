package qp_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/qp/go"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {

	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", log.Llongfile|log.Ldate|log.Lmicroseconds)

	var l qp.Logger
	l = qp.LogLogger(logger)

	l.Error("o", "n", "e")
	l.Errorf("t%s", "wo")
	l.Errorf("%s", "three")

	require.Contains(t, buf.String(), "one")
	require.Contains(t, buf.String(), "two")
	require.Contains(t, buf.String(), "three")
	require.Contains(t, buf.String(), "logger_test.go")
	require.Equal(t, 4, len(strings.Split(buf.String(), "\n")))

}

type logger struct {
	output []string
}

func (l *logger) Error(v ...interface{}) {
	l.output = append(l.output, fmt.Sprint(v...))
}
func (l *logger) Errorf(format string, v ...interface{}) {
	l.output = append(l.output, fmt.Sprintf(format, v...))
}
func (l *logger) Info(v ...interface{}) {
	l.output = append(l.output, fmt.Sprint(v...))
}
func (l *logger) Infof(format string, v ...interface{}) {
	l.output = append(l.output, fmt.Sprintf(format, v...))
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
	ls.Errorf("(%s)", "three")

	require.Equal(t, 3, len(l1.output))
	require.Equal(t, 3, len(l2.output))
	require.Equal(t, 3, len(l3.output))

	require.Equal(t, "one", l1.output[0])
	require.Equal(t, "two", l1.output[1])
	require.Equal(t, "(three)", l1.output[2])

	require.Equal(t, "one", l2.output[0])
	require.Equal(t, "two", l2.output[1])
	require.Equal(t, "(three)", l2.output[2])

	require.Equal(t, "one", l3.output[0])
	require.Equal(t, "two", l3.output[1])
	require.Equal(t, "(three)", l3.output[2])

}
