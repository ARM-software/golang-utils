package summary

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestSummaryLoggerInMemory(t *testing.T) {
	logger, err := NewInMemorySummaryLogger("summary-test")
	require.NoError(t, err)
	require.NoError(t, logger.SetLogSource("build"))

	logger.Log("hello")
	logger.LogError("boom")
	require.NoError(t, logger.WriteString("## Raw"))
	require.NoError(t, logger.WriteStringF(" %s", "section"))
	require.NoError(t, logger.WriteLine("line"))
	require.NoError(t, logger.WriteLineF("- %s", "formatted"))
	require.NoError(t, logger.Flush())

	expected := "[build] hello" + platform.LineSeparator() +
		"[build] ERROR: boom" + platform.LineSeparator() +
		"## Raw sectionline" + platform.LineSeparator() +
		"- formatted" + platform.LineSeparator()
	assert.Equal(t, expected, logger.Content())

	require.NoError(t, logger.Clear())
	assert.Empty(t, logger.Content())
}

func TestFileSummaryLogger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	logger, err := NewFileSummaryLogger(path, "summary-test")
	require.NoError(t, err)
	require.NoError(t, logger.SetLogSource("build"))

	logger.Log("hello")
	logger.LogError("boom")
	require.NoError(t, logger.WriteLine("## Raw"))
	require.NoError(t, logger.WriteLineF("- %s", "formatted"))
	require.NoError(t, logger.Flush())

	content, err := filesystem.ReadFile(path)
	require.NoError(t, err)
	expected := "[build] hello" + platform.LineSeparator() +
		"[build] ERROR: boom" + platform.LineSeparator() +
		"## Raw" + platform.LineSeparator() +
		"- formatted" + platform.LineSeparator()
	assert.Equal(t, expected, string(content))

	require.NoError(t, logger.Clear())
	fileInfo, err := filesystem.Stat(path)
	require.NoError(t, err)
	assert.Zero(t, fileInfo.Size())
	assert.Empty(t, logger.Content())

	require.NoError(t, logger.Close())
}

func TestGitHubSummaryLoggerFromEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv(GitHubStepSummaryEnvironmentVariable, path)

	logger, err := NewGitHubSummaryLoggerFromEnvironment("summary-test")
	require.NoError(t, err)
	require.NoError(t, logger.WriteLine("hello"))
	require.NoError(t, logger.Close())

	content, err := filesystem.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello"+platform.LineSeparator(), string(content))

	t.Setenv(GitHubStepSummaryEnvironmentVariable, "")
	_, err = NewGitHubSummaryLoggerFromEnvironment("summary-test")
	errortest.AssertError(t, err, commonerrors.ErrUndefined)
}

func TestSummaryLoggerValidation(t *testing.T) {
	logger := &SummaryLogger{}
	errortest.AssertError(t, logger.Check(), commonerrors.ErrNoLogger)
}
