package logger

import (
	stdlog "log"
)

type Logger struct {
	prefix string
}

func New(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

func (l *Logger) Infof(format string, args ...any) {
	stdlog.Printf("["+l.prefix+"] "+format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	stdlog.Printf("["+l.prefix+"] ERROR: "+format, args...)
}
