package logrimp

import (
	"github.com/go-logr/logr"
	"github.com/zailic/slogr"
	"golang.org/x/exp/slog"
)

// NewSlogLogger returns a new [slog logger](see https://pkg.go.dev/golang.org/x/exp/slog) which will be part of the standard library.
func NewSlogLogger(logger *slog.Logger) logr.Logger {
	// FIXME change dependency when needed https://github.com/go-logr/logr/issues/171
	return slogr.New(logger)
}
