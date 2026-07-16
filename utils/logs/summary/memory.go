package summary

import baselogs "github.com/ARM-software/golang-utils/utils/logs"

var _ ISummaryLogger = &InMemorySummaryLogger{}

// NewInMemorySummaryLogger creates an in-memory summary logger backed by the
// repository's plain string logger.
func NewInMemorySummaryLogger(loggerSource string) (logger *InMemorySummaryLogger, err error) {
	bLogger, err := baselogs.NewPlainStringLogger()
	if err != nil {
		return
	}
	err = bLogger.SetLoggerSource(loggerSource)
	if err != nil {
		return
	}
	logger = &InMemorySummaryLogger{baseSummaryLogger{
		Loggers: bLogger,
	}}
	return
}

// InMemorySummaryLogger stores summary content in memory.
type InMemorySummaryLogger struct {
	baseSummaryLogger
}

// GetSummary returns the accumulated summary content.
func (s *InMemorySummaryLogger) GetSummary() string {
	if l, ok := s.Loggers.(*baselogs.StringLoggers); ok {
		return l.GetLogContent()
	}
	return ""
}
