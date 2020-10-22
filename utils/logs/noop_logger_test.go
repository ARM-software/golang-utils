package logs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopLogger(t *testing.T) {
	loggers, err := NewNoopLogger("Test")
	require.Nil(t, err)
	_testLog(t, loggers)
}
