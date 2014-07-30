package qp

// Logger represents a type capable of logging errors.
type Logger interface {
	// Printf prints to the log.
	// Arguments are handled in the manner of fmt.Print.
	Printf(format string, v ...interface{})
	// Print prints to the log.
	// Arguments are handled in the manner of fmt.Print.
	Print(v ...interface{})
	// Println prints to the log.
	// Arguments are handled in the manner of fmt.Print.
	Println(v ...interface{})
}

// Loggers funnels many loggers into one logger.
// Calls to this logger will be passed to each of the specified
// Loggers.
func Loggers(l ...Logger) Logger {
	lgs := make(loggers, len(l))
	for i, logger := range l {
		lgs[i] = logger
	}
	return lgs
}

type loggers []Logger

func (l loggers) Printf(format string, v ...interface{}) {
	for _, logger := range l {
		logger.Printf(format, v...)
	}
}
func (l loggers) Print(v ...interface{}) {
	for _, logger := range l {
		logger.Print(v...)
	}
}
func (l loggers) Println(v ...interface{}) {
	for _, logger := range l {
		logger.Println(v...)
	}
}

// NilLogger represents a Logger that ignores all calls.
var NilLogger Logger = nilLoggerI

// nilLoggerI is the nil instance of nilLogger
var nilLoggerI *nilLogger

type nilLogger struct{}

func (l *nilLogger) Printf(format string, v ...interface{}) {}
func (l *nilLogger) Print(v ...interface{})                 {}
func (l *nilLogger) Println(v ...interface{})               {}
