package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStdLogger(t *testing.T) {
	loggers, err := CreateStdLogger("Test")
	require.Nil(t, err)
	_testLog(t, loggers)
}

func TestAsynchronousStdLogger(t *testing.T) {
	loggers, err := NewAsynchronousStdLogger("Test", 1024, 2*time.Millisecond, "test source")
	require.Nil(t, err)
	_testLog(t, loggers)
}
