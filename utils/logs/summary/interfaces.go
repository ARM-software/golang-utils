package summary

import (
	"io"

	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

const (
	// GitHubStepSummaryEnvironmentVariable is the environment variable pointing to
	// the current step's GitHub Actions summary file.
	GitHubStepSummaryEnvironmentVariable = "GITHUB_STEP_SUMMARY"
)

// Writer writes Markdown summary content to an underlying summary sink.
type Writer interface {
	io.Closer
	// WriteMarkdown appends Markdown content to the summary sink.
	WriteMarkdown(markdown string) error
	// Clear removes any currently accumulated summary content for the sink.
	Clear() error
}

// Loggers extends [logs.Loggers] with summary-specific operations.
type Loggers interface {
	baselogs.Loggers
	// WriteMarkdown appends raw Markdown content to the summary output.
	WriteMarkdown(markdown string) error
	// Clear removes any currently accumulated summary content.
	Clear() error
}
