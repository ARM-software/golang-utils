package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogMessage(t *testing.T) {
	loggers, err := NewJSONLogger(&StdWriter{}, "Test", "TestLogMessage")
	require.Nil(t, err)
	_testLog(t, loggers)
}

func TestLogMessageToSlowLogger(t *testing.T) {
	stdloggers, err := CreateStdLogger("ERR:")
	require.Nil(t, err)
	loggers, err := NewJSONLoggerForSlowWriter(&SlowWriter{}, 1024, 2*time.Millisecond, "Test", "TestLogMessageToSlowLogger", stdloggers)
	require.Nil(t, err)
	_testLog(t, loggers)
	time.Sleep(100 * time.Millisecond)
}
