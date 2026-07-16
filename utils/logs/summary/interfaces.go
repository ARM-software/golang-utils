package summary

import (
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

//go:generate go tool mockgen -source=./interfaces.go -destination=../../mocks/mock_summary.go -package=mocks

// ISummaryLogger extends [logs.Loggers] with operations for appending
// human-readable summary content.
//
// The summary API writes plain strings. Implementations may render or persist
// those strings in richer formats such as Markdown.
type ISummaryLogger interface {
	baselogs.Loggers
	// WriteString appends summary content and ensures it ends with a newline.
	WriteString(content string) error
	// WriteStringF appends formatted summary content and ensures it ends with a
	// newline.
	WriteStringF(format string, values ...any) error
}
