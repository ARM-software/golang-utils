package summary

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestGitHubWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	writer, err := NewGitHubWriter(path)
	require.NoError(t, err)

	require.NoError(t, writer.WriteMarkdown("## Summary\n"))
	require.NoError(t, writer.WriteMarkdown("- entry\n"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "## Summary\n- entry\n", string(content))

	require.NoError(t, writer.Clear())
	content, err = os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, string(content))
}

func TestGitHubWriterFromEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv(GitHubStepSummaryEnvironmentVariable, path)

	writer, err := NewGitHubWriterFromEnvironment()
	require.NoError(t, err)
	require.NoError(t, writer.WriteMarkdown("hello\n"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(content))

	t.Setenv(GitHubStepSummaryEnvironmentVariable, "")
	_, err = NewGitHubWriterFromEnvironment()
	errortest.AssertError(t, err, commonerrors.ErrUndefined)
}

func TestSummaryLogger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	logger, err := NewGitHubLogger(path, "summary-test")
	require.NoError(t, err)
	require.NoError(t, logger.SetLogSource("build"))

	logger.Log("hello")
	logger.LogError("boom")
	require.NoError(t, logger.WriteMarkdown("## Raw\n"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "[build] hello\n[build] ERROR: boom\n## Raw\n", string(content))

	require.NoError(t, logger.Clear())
	content, err = os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, string(content))
}

func TestSummaryLoggerValidation(t *testing.T) {
	_, err := NewLogger(nil, "summary-test")
	errortest.AssertError(t, err, commonerrors.ErrNoLogger)
}
