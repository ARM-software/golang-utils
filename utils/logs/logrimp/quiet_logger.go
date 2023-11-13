package logrimp

import (
	"github.com/go-logr/logr"
)

type quietLogger struct {
	logger logr.Logger
}

func (l *quietLogger) Init(_ logr.RuntimeInfo) {
	// ignored.
}

func (l *quietLogger) Enabled(int) bool {
	return false
}

func (l *quietLogger) Info(_ int, _ string, _ ...any) {
	// Ignored.
}

func (l *quietLogger) Error(err error, msg string, keysAndValues ...any) {
	l.logger.Error(err, msg, keysAndValues...)
}

func (l *quietLogger) WithValues(keysAndValues ...any) logr.LogSink {
	l.logger.WithValues(keysAndValues...)
	return l
}

func (l *quietLogger) WithName(name string) logr.LogSink {
	l.logger.WithName(name)
	return l
}

// NewQuietLogger returns a quiet logger which only logs errors.
func NewQuietLogger(logger logr.Logger) logr.Logger {
	return logr.New(&quietLogger{logger: logger})
}
