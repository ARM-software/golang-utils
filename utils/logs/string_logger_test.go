package logs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringLogger(t *testing.T) {
	loggers, err := CreateStringLogger("Test")
	require.Nil(t, err)
	_testLog(t, loggers)
	loggers.LogError("Test err")
	loggers.Log("Test1")
	contents := loggers.GetLogContent()
	require.NotZero(t, contents)
	require.True(t, strings.Contains(contents, "Test err"))
	require.True(t, strings.Contains(contents, "Test1"))
}
