package summary

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestSummaryLoggerInMemory(t *testing.T) {
	defer goleak.VerifyNone(t)
	logger, err := NewInMemorySummaryLogger("summary-test")
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()
	require.NoError(t, logger.SetLogSource("build"))
	expected := generateSummary(t, logger)
	assert.Equal(t, expected, logger.GetSummary())
	require.NoError(t, logger.Close())
}

func TestFileSummaryLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	path := filepath.Join(t.TempDir(), "summary.md")
	logger, err := NewFileSummaryLogger(path, "summary-test")
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()
	expected := generateSummary(t, logger)
	content, err := filesystem.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expected, string(content))
}

func generateSummary(t *testing.T, summaryLogger ISummaryLogger) (expected string) {
	t.Helper()
	require.NotNil(t, summaryLogger)
	paragraph := faker.Paragraph()

	summaryLogger.Log("hello")
	summaryLogger.LogError("boom")
	require.NoError(t, summaryLogger.WriteString(""))
	require.NoError(t, summaryLogger.WriteString("## Raw"))
	require.NoError(t, summaryLogger.WriteStringF(" %s", "section"))
	require.NoError(t, summaryLogger.WriteString(paragraph))
	require.NoError(t, summaryLogger.WriteStringF("- %s", "formatted"))

	expected = fmt.Sprintf("hello\nboom\n\n## Raw\n section\n%v\n- formatted\n", paragraph)
	return
}
