package summary

import (
	"fmt"
	"strings"

	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

var _ ISummaryLogger = &baseSummaryLogger{}

// baseSummaryLogger adapts a standard logger pair to the summary logger
// interface.
type baseSummaryLogger struct {
	baselogs.Loggers
}

// WriteString appends summary content and ensures it ends with a newline.
func (b *baseSummaryLogger) WriteString(content string) error {
	return b.WriteStringF("%v", content)
}

// WriteStringF appends formatted summary content and ensures it ends with a
// newline.
func (b *baseSummaryLogger) WriteStringF(format string, values ...any) error {
	b.Log(strings.TrimRight(fmt.Sprintf(format, values...), "\r\n"))
	return nil
}
