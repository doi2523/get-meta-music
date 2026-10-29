// Package logger provides internal logging utilities for get-music-metadata.
// This package is placed in the internal directory so it cannot be imported externally.
package logger

import (
	"fmt"
	"log"
	"sync"
)

// Logger defines the minimal logging interface.
type Logger interface {
	Printf(format string, v ...interface{})
}

type nopLogger struct{}

func (n *nopLogger) Printf(format string, v ...interface{}) {}

// NopLogger is a no-op logger that discards all log messages.
var NopLogger Logger = &nopLogger{}

var (
	mu      sync.RWMutex
	current Logger = log.Default()
)

// SetLogger sets the active logger. Pass NopLogger to silence logs completely.
func SetLogger(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	if l == nil {
		current = NopLogger
	} else {
		current = l
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

// Printf routes formatted messages to the configured logger.
func Printf(format string, v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if current != nil {
		current.Printf(format, v...)
	}
}

// Println routes unformatted messages to the configured logger.
func Println(v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if current != nil {
		current.Printf("%s", fmt.Sprint(v...))
	}
}
