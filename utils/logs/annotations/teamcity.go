package annotations

import (
	"fmt"
	"strings"

	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// TeamCityFormatter formats annotations as TeamCity service messages.
//
// Reference:
//   - https://www.jetbrains.com/help/teamcity/service-messages.html
type TeamCityFormatter struct{}

// NewTeamCityLogger creates a TeamCity-formatted annotation logger backed by baseLogger.
func NewTeamCityLogger(baseLogger baselogs.Loggers) (*AnnotationLogger, error) {
	return NewLogger(baseLogger, TeamCityFormatter{})
}

func (TeamCityFormatter) Format(annotation *Annotation) string {
	if annotation == nil {
		return ""
	}
	text := annotation.Message
	if reflection.IsNotEmpty(annotation.File) {
		text = fmt.Sprintf("%s (%s)", annotation.Message, annotation.File)
	}
	switch annotation.Severity {
	case SeverityError:
		return fmt.Sprintf("##teamcity[buildProblem description='%s']", escapeTeamCity(text))
	case SeverityWarning:
		return fmt.Sprintf("##teamcity[message text='%s' status='WARNING']", escapeTeamCity(text))
	default:
		return fmt.Sprintf("##teamcity[message text='%s' status='NORMAL']", escapeTeamCity(text))
	}
}

func escapeTeamCity(value string) string {
	return strings.NewReplacer(
		"|", "||",
		"'", "|'",
		"\n", "|n",
		"\r", "|r",
		"[", "|[",
		"]", "|]",
	).Replace(value)
}
