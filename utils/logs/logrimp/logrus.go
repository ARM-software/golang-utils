package logrimp

import (
	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// NewLogrusLogger returns a logrus logger.
func NewLogrusLogger(logger logrus.FieldLogger) logr.Logger {
	return logrusr.New(logger)
}
