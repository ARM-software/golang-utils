package annotations

import (
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

//go:generate go tool mockgen -source=./interfaces.go -destination=../../mocks/mock_annotations.go -package=mocks

// IAnnotationLogger extends [logs.Loggers] with methods for writing structured
// annotations.
//
// An annotation logger emits issue records in a format that CI systems can
// interpret specially, for example by surfacing them as file-linked errors or
// warnings in the build UI.
type IAnnotationLogger interface {
	baselogs.Loggers
	// WriteAnnotation writes annotation using the configured formatter.
	WriteAnnotation(annotation *Annotation) error
	// WriteError writes an error-level annotation.
	WriteError(message string, options ...AnnotationOption) error
	// WriteWarning writes a warning-level annotation.
	WriteWarning(message string, options ...AnnotationOption) error
	// WriteNotice writes a notice-level annotation.
	WriteNotice(message string, options ...AnnotationOption) error
}

// IFormatter converts an annotation into the platform-specific line to emit.
//
// References:
//   - GitHub Actions workflow commands:
//     https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands
//   - Azure DevOps logging commands:
//     https://learn.microsoft.com/en-us/azure/devops/pipelines/scripts/logging-commands
//   - TeamCity service messages:
//     https://www.jetbrains.com/help/teamcity/service-messages.html
type IFormatter interface {
	Format(annotation *Annotation) string
}
