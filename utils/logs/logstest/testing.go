package logstest

import (
	"testing"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	logrusTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/ARM-software/golang-utils/utils/logs/logrimp"
)

// NewNullTestLogger returns a logger to nothing
func NewNullTestLogger() logr.Logger {
	// FIXME should probably be replaced by NoOp
	internalLogger, _ := logrusTest.NewNullLogger()
	return logrusr.New(internalLogger)
}

// NewStdTestLogger returns a test logger to standard output.
func NewStdTestLogger() logr.Logger {
	return logrimp.NewStdOutLogr()
}

// NewTestLogger returns a logger to use in tests
func NewTestLogger(t *testing.T) logr.Logger {
	return testr.New(t)
}
