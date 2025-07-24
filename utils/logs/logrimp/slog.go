package logrimp

import (
	"log/slog"

	"github.com/go-logr/logr"
)

// NewSlogLogger returns a new [slog logger](see https://pkg.go.dev/golang.org/x/exp/slog) which will be part of the standard library.
func NewSlogLogger(logger *slog.Logger) logr.Logger {
	if logger == nil {
		return logr.Discard()
	}
	return logr.FromSlogHandler(logger.Handler())
}
