package logrimp

import (
	"github.com/evanphx/hclogr"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
)

// NewHclogLogger returns a new HCLog logger.
func NewHclogLogger(logger hclog.Logger) logr.Logger {
	return hclogr.Wrap(logger)
}
