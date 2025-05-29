package logs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFIFOLogger(t *testing.T) {
	loggers, err := NewFIFOLogger("Test")
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError("Test err")
	loggers.Log("Test1")
	contents := loggers.Read()
	require.NotEmpty(t, contents)
	require.True(t, strings.Contains(contents, "Test err"))
	require.True(t, strings.Contains(contents, "Test1"))
	loggers.Log("Test2")
	contents = loggers.Read()
	require.NotEmpty(t, contents)
	require.False(t, strings.Contains(contents, "Test err"))
	require.False(t, strings.Contains(contents, "Test1"))
	require.True(t, strings.Contains(contents, "Test2"))
	contents = loggers.Read()
	require.Empty(t, contents)
}

func TestPlainFIFOLogger(t *testing.T) {
	loggers, err := NewPlainFIFOLogger()
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError("Test err")
	loggers.Log("Test1")
	contents := loggers.Read()
	require.NotEmpty(t, contents)
	require.True(t, strings.Contains(contents, "Test err"))
	require.True(t, strings.Contains(contents, "Test1"))
	loggers.Log("Test2")
	contents = loggers.Read()
	require.NotEmpty(t, contents)
	require.False(t, strings.Contains(contents, "Test err"))
	require.False(t, strings.Contains(contents, "Test1"))
	require.True(t, strings.Contains(contents, "Test2"))
	contents = loggers.Read()
	require.Empty(t, contents)
}
