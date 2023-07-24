package logrimp

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// NewZapLogger returns a new zap logger
func NewZapLogger(logger *zap.Logger) logr.Logger {
	return zapr.NewLogger(logger)
}
