package summary

import (
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

// ISummaryLogger extends [logs.Loggers] with summary-specific operations.
type ISummaryLogger interface {
	baselogs.Loggers
	// WriteString appends raw summary content exactly as supplied. It does not add
	// an end-of-line marker automatically. Destinations may render the string
	// using Markdown syntax.
	WriteString(content string) error
	// WriteStringF appends formatted summary content exactly as supplied. It does
	// not add an end-of-line marker automatically.
	WriteStringF(format string, values ...any) error
	// WriteLine appends summary content and ensures it ends with a newline.
	WriteLine(content string) error
	// WriteLineF appends formatted summary content and ensures it ends with a
	// newline.
	WriteLineF(format string, values ...any) error
}



