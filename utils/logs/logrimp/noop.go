package logrimp

import "github.com/go-logr/logr"

// NewNoopLogger returns a discarding.
func NewNoopLogger() logr.Logger {
	return logr.Discard()
}
