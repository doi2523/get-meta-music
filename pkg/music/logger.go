package music

import (
	"fmt"
	"log"
	"sync"
)

// Logger defines the minimal logging interface used throughout the music package.
type Logger interface {
	Printf(format string, v ...interface{})
}

type nopLogger struct{}

func (n *nopLogger) Printf(format string, v ...interface{}) {}

// NopLogger is a no-op logger that discards all log messages.
var NopLogger Logger = &nopLogger{}

var (
	loggerMu sync.RWMutex
	logger   Logger = log.Default()
)

// SetLogger sets the active logger. Pass NopLogger to silence logs completely.
func SetLogger(l Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if l == nil {
		logger = NopLogger
	} else {
		logger = l
	}
}

// SetSilent controls whether internal logging is silenced or restored to log.Default().
func SetSilent(silent bool) {
	if silent {
		SetLogger(NopLogger)
	} else {
		SetLogger(log.Default())
	}
}

// logPrintf routes formatted messages to the configured logger.
func logPrintf(format string, v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if logger != nil {
		logger.Printf(format, v...)
	}
}

// logPrintln routes unformatted messages to the configured logger.
func logPrintln(v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if logger != nil {
		logger.Printf("%s", fmt.Sprint(v...))
	}
}
