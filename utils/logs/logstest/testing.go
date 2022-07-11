package logstest

import (
	"testing"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/ARM-software/golang-utils/utils/reflection"
)

// NewNullTestLogger returns a logger to nothing
func NewNullTestLogger() logr.Logger {
	internalLogger, _ := logrusTest.NewNullLogger()
	return logrusr.NewLogger(internalLogger)
}

// NewStdTestLogger returns a test logger to standard output.
func NewStdTestLogger() logr.Logger {
	return logrusr.NewLogger(logrus.New())
}

// NewTestLogger returns a logger to use in tests
func NewTestLogger(t *testing.T) logr.Logger {
	return &testLogger{t: t}
}

type testLogger struct {
	t *testing.T
}

func (t *testLogger) Enabled() bool {
	return true
}

func (t *testLogger) Info(msg string, keysAndValues ...interface{}) {
	if reflection.IsEmpty(keysAndValues) {
		t.t.Log(msg)
	} else {
		t.t.Logf("%s -- %v", msg, keysAndValues)
	}
}

func (t *testLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if reflection.IsEmpty(keysAndValues) {
		t.t.Logf("%s: %v", msg, err)
	} else {
		t.t.Logf("%s: %v -- %v", msg, err, keysAndValues)
	}
}

func (t *testLogger) V(level int) logr.Logger {
	return t
}

func (t *testLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return t
}

func (t *testLogger) WithName(name string) logr.Logger {
	return t
}
