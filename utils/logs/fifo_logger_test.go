package logs

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFIFOLoggerRead(t *testing.T) {
	loggers, err := NewFIFOLogger()
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError("Test err")
	loggers.Log("Test1")
	contents := loggers.Read()
	require.NotEmpty(t, contents)
	time.Sleep(200 * time.Millisecond) // account for slow polling
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

func TestPlainFIFOLoggerRead(t *testing.T) {
	loggers, err := NewPlainFIFOLogger()
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError("Test err")
	loggers.Log("Test1")
	time.Sleep(200 * time.Millisecond) // account for slow polling
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

func TestFIFOLoggerReadlines(t *testing.T) {
	loggers, err := NewFIFOLogger()
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError("Test err\n")
	loggers.Log("Test1")
	count := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var b strings.Builder
	for line := range loggers.ReadLines(ctx) {
		_, err := b.WriteString(line + "\n")
		require.NoError(t, err)
		count++
	}

	assert.Regexp(t, regexp.MustCompile(`\[Test\] Error: .* .* Test err\n\[Test\] Output: .* .* Test1\n`), b.String())
	assert.Equal(t, 2, count)
}

func TestPlainFIFOLoggerReadlines(t *testing.T) {
	loggers, err := NewPlainFIFOLogger()
	require.NoError(t, err)
	testLog(t, loggers)

	go func() {
		time.Sleep(500 * time.Millisecond)
		loggers.LogError("Test err")
		loggers.Log("")
		time.Sleep(100 * time.Millisecond)
		loggers.Log("Test1")
		loggers.Log("\n\n\n")
		time.Sleep(200 * time.Millisecond)
		loggers.Log("Test2\n")
	}()

	count := 0
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var b strings.Builder
	for line := range loggers.ReadLines(ctx) {
		_, err := b.WriteString(line + "\n")
		require.NoError(t, err)
		count++
	}

	assert.Equal(t, "Test errTest1\nTest2\n", b.String())
	assert.Equal(t, 3, count)
}
