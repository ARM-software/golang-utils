package logstest

import (
	"testing"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

// NewNullTestLogger returns a logger to nothing
func NewNullTestLogger() logr.Logger {
	internalLogger, _ := logrusTest.NewNullLogger()
	return logrusr.New(internalLogger)
}

// NewStdTestLogger returns a test logger to standard output.
func NewStdTestLogger() logr.Logger {
	return logrusr.New(logrus.New())
}

// NewTestLogger returns a logger to use in tests
func NewTestLogger(t *testing.T) logr.Logger {
	return testr.New(t)
}
