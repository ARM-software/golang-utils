package annotations

import (
	"fmt"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// GitHubFormatter formats annotations as GitHub Actions annotation commands.
//
// Reference:
//   - https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands
type GitHubFormatter struct{}

// NewGitHubLogger creates a GitHub-formatted annotation logger backed by baseLogger.
func NewGitHubLogger(baseLogger baselogs.Loggers) (*AnnotationLogger, error) {
	return NewLogger(baseLogger, GitHubFormatter{})
}

func (GitHubFormatter) Format(annotation *Annotation) string {
	if annotation == nil {
		return ""
	}
	command := strings.ToLower(strings.TrimPrefix(annotation.Severity.String(), "Severity"))
	properties := collection.Filter([]string{
		func() string {
			if reflection.IsEmpty(annotation.File) {
				return ""
			}
			return fmt.Sprintf("file=%s", escapeGitHubProperty(annotation.File))
		}(),
		func() string {
			if annotation.Line == nil {
				return ""
			}
			return fmt.Sprintf("line=%d", *annotation.Line)
		}(),
		func() string {
			if annotation.Column == nil {
				return ""
			}
			return fmt.Sprintf("col=%d", *annotation.Column)
		}(),
	}, func(value string) bool {
		return reflection.IsNotEmpty(value)
	})
	propertyBlock := ""
	if len(properties) > 0 {
		propertyBlock = " " + strings.Join(properties, ",")
	}
	return fmt.Sprintf("::%s%s::%s", command, propertyBlock, escapeGitHubMessage(annotation.Message))
}

func escapeGitHubMessage(value string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(value)
}

func escapeGitHubProperty(value string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C").Replace(value)
}
