package annotations

import (
	"fmt"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// AzureDevOpsFormatter formats annotations as Azure DevOps logging commands.
//
// Reference:
//   - https://learn.microsoft.com/en-us/azure/devops/pipelines/scripts/logging-commands
type AzureDevOpsFormatter struct{}

// NewAzureDevOpsLogger creates an Azure DevOps-formatted annotation logger backed by baseLogger.
func NewAzureDevOpsLogger(baseLogger baselogs.Loggers) (*AnnotationLogger, error) {
	return NewLogger(baseLogger, AzureDevOpsFormatter{})
}

func (AzureDevOpsFormatter) Format(annotation *Annotation) string {
	if annotation == nil {
		return ""
	}
	typeName := strings.ToLower(strings.TrimPrefix(annotation.Severity.String(), "Severity"))
	if annotation.Severity == SeverityNotice {
		typeName = "warning"
	}
	properties := collection.Filter([]string{
		fmt.Sprintf("type=%s", typeName),
		func() string {
			if reflection.IsEmpty(annotation.File) {
				return ""
			}
			return fmt.Sprintf("sourcepath=%s", escapeAzureProperty(annotation.File))
		}(),
		func() string {
			if annotation.Line == nil {
				return ""
			}
			return fmt.Sprintf("linenumber=%d", *annotation.Line)
		}(),
		func() string {
			if annotation.Column == nil {
				return ""
			}
			return fmt.Sprintf("columnnumber=%d", *annotation.Column)
		}(),
	}, func(value string) bool {
		return reflection.IsNotEmpty(value)
	})
	return fmt.Sprintf("##vso[task.logissue %s]%s", strings.Join(properties, ";"), escapeAzureMessage(annotation.Message))
}

func escapeAzureMessage(value string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(value)
}

func escapeAzureProperty(value string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ";", "%3B", "]", "%5D").Replace(value)
}
