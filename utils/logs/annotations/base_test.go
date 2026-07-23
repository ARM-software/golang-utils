package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

func TestAnnotationLoggerFromLoggers(t *testing.T) {
	base, err := baselogs.NewPlainStringLogger()
	require.NoError(t, err)

	logger, err := NewGitHubLogger(base)
	require.NoError(t, err)
	line := 3
	annotation := newAnnotation(SeverityError, "broken", WithFile("pkg/file.go"), WithLine(line))
	require.NoError(t, logger.WriteAnnotation(&annotation))
	require.NoError(t, logger.WriteWarning("warn"))
	require.NoError(t, logger.WriteNotice("note"))

	assert.Equal(t,
		"::error file=pkg/file.go,line=3::broken\n::warning::warn\n::notice::note\n",
		base.GetLogContent(),
	)
}
