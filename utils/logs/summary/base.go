package summary

import (
	"fmt"
	"strings"

	baselogs "github.com/ARM-software/golang-utils/utils/logs"
)

var _ ISummaryLogger = &baseSummaryLogger{}

type baseSummaryLogger struct {
	baselogs.Loggers
}

func (b *baseSummaryLogger) WriteString(content string) error {
	return b.WriteStringF("%v", content)
}

func (b *baseSummaryLogger) WriteStringF(format string, values ...any) error {
	b.Loggers.Log(fmt.Sprintf(format, values...))
	return nil
}

func (b *baseSummaryLogger) WriteLine(content string) error {
	return b.WriteLineF("%v", content)
}

func (b *baseSummaryLogger) WriteLineF(format string, values ...any) error {
	b.Loggers.Log(strings.TrimRight(fmt.Sprintf(format, values...), "\r\n") + "\n")
	return nil
}
