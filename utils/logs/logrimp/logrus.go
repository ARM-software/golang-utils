package logrimp

import (
	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// NewLogrusLogger returns a logrus logger.
func NewLogrusLogger(logger logrus.FieldLogger, opts ...logrusr.Option) logr.Logger {
	return logrusr.New(logger, opts...)
}
