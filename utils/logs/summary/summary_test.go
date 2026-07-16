package summary

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestSummaryLoggerInMemory(t *testing.T) {
	defer goleak.VerifyNone(t)
	logger, err := NewInMemorySummaryLogger("summary-test")
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()
	require.NoError(t, logger.SetLogSource("build"))

	logger.Log("hello")
	logger.LogError("boom")
	require.NoError(t, logger.WriteString("## Raw"))
	require.NoError(t, logger.WriteStringF(" %s", "section"))
	require.NoError(t, logger.WriteLine("line"))
	require.NoError(t, logger.WriteLineF("- %s", "formatted"))

	expected := "[build] hello" + platform.LineSeparator() +
		"[build] ERROR: boom" + platform.LineSeparator() +
		"## Raw sectionline" + platform.LineSeparator() +
		"- formatted" + platform.LineSeparator()
	assert.Equal(t, expected, logger.GetSummary())
	require.NoError(t, logger.Close())
}

func TestFileSummaryLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	path := filepath.Join(t.TempDir(), "summary.md")
	logger, err := NewFileSummaryLogger(path, "summary-test")
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()
	require.NoError(t, logger.SetLogSource("build"))

	logger.Log("hello")
	logger.LogError("boom")
	require.NoError(t, logger.WriteLine("## Raw"))
	require.NoError(t, logger.WriteLineF("- %s", "formatted"))
	require.NoError(t, logger.Close())

	content, err := filesystem.ReadFile(path)
	require.NoError(t, err)
	expected := "[build] hello" + platform.LineSeparator() +
		"[build] ERROR: boom" + platform.LineSeparator() +
		"## Raw" + platform.LineSeparator() +
		"- formatted" + platform.LineSeparator()
	assert.Equal(t, expected, string(content))
}
