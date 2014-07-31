package qp

import (
	"fmt"
	"log"
)

// Logger represents a type capable of logging info, warnings and errors.
type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

type nilLogger struct{}

func (n *nilLogger) Error(...interface{})          {}
func (n *nilLogger) Errorf(string, ...interface{}) {}

// NilLogger represents a Logger the calls to which are
// disregarded.
var NilLogger *nilLogger

type loggers []Logger

var _ Logger = (loggers)(nil)

// Loggers creates a single Logger from many other Logger
// objects.
func Loggers(ls ...Logger) Logger {
	l := make(loggers, len(ls))
	for i, logger := range ls {
		l[i] = logger
	}
	return l
}

func (ls loggers) Error(args ...interface{}) {
	for _, l := range ls {
		l.Error(args...)
	}
}
func (ls loggers) Errorf(format string, args ...interface{}) {
	ls.Error(fmt.Sprintf(format, args...))
}

type loglogger struct {
	logger *log.Logger
}

func (l *loglogger) output(s string) {
	l.logger.Output(3, s)
}
func (l *loglogger) Error(args ...interface{}) {
	l.output(fmt.Sprint(args...))
}
func (l *loglogger) Errorf(format string, args ...interface{}) {
	l.output(fmt.Sprintf(format, args...))
}

// LogLogger creates a Logger that logs to the specified
// log.Logger.
func LogLogger(logger *log.Logger) Logger {
	return &loglogger{logger: logger}
}
