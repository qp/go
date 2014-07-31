package qp

// Logger represents a type capable of logging info, warnings and errors.
type Logger interface {
	Error(...interface{})
}

type nilLogger struct{}

func (n *nilLogger) Error(...interface{}) {}

var NilLogger *nilLogger

type loggers []Logger

var _ Logger = (loggers)(nil)

func Loggers(ls ...Logger) Logger {
	l := make(loggers, len(ls))
	for i, logger := range ls {
		l[i] = logger
	}
	return l
}

func (ls loggers) Error(a ...interface{}) {
	for _, l := range ls {
		l.Error(a...)
	}
}
