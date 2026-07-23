package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

func TestTeamCityFormatter(t *testing.T) {
	formatter := TeamCityFormatter{}
	annotation := &Annotation{Severity: SeverityError, Message: "broken", File: "pkg/file.go"}
	assert.Equal(t,
		"##teamcity[buildProblem description='broken (pkg/file.go)']",
		formatter.Format(annotation),
	)
}

func TestTeamCityLoggerFromLoggers(t *testing.T) {
	base, err := baselogs.NewPlainStringLogger()
	require.NoError(t, err)

	logger, err := NewTeamCityLogger(base)
	require.NoError(t, err)
	require.NoError(t, logger.WriteWarning("watch this"))

	assert.Equal(t, "##teamcity[message text='watch this' status='WARNING']\n", base.GetLogContent())
}
