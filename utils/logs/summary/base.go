package summary

import (
	"fmt"
	"strings"

	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

var _ ISummaryLogger = &SummaryLogger{}

// SummaryLogger adapts a regular [logs.Loggers] implementation to the
// [ISummaryLogger] interface.
//
// It normalises each summary write to a single trailing newline before sending
// the content through the wrapped logger.
type SummaryLogger struct {
	baselogs.Loggers
}

// WriteString appends summary content and ensures it ends with a newline.
func (b *SummaryLogger) WriteString(content string) error {
	return b.WriteStringF("%v", content)
}

// WriteStringF appends formatted summary content and ensures it ends with a
// newline.
//
// Any trailing CR/LF characters already present in the formatted text are
// trimmed so the wrapped logger emits one logical summary line per call.
func (b *SummaryLogger) WriteStringF(format string, values ...any) error {
	b.Log(strings.TrimRight(fmt.Sprintf(format, values...), "\r\n"))
	return nil
}

// NewSummaryLogger creates a summary logger backed by baseLogger.
func NewSummaryLogger(baseLogger baselogs.Loggers) (logger *SummaryLogger) {
	return &SummaryLogger{Loggers: baseLogger}
}
