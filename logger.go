package qp

// Logger represents a type capable of logging errors.
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
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
