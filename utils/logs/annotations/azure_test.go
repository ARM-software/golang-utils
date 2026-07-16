package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzureDevOpsFormatter(t *testing.T) {
	formatter := AzureDevOpsFormatter{}
	line := 8
	annotation := &Annotation{Severity: SeverityWarning, Message: "problem", File: "pkg/file.go", Line: &line}
	assert.Equal(t,
		"##vso[task.logissue type=warning;sourcepath=pkg/file.go;linenumber=8]problem",
		formatter.Format(annotation),
	)
}
