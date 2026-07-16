package summary

import baselogs "github.com/ARM-software/golang-utils/utils/logs"

var _ ISummaryLogger = &InMemorySummaryLogger{}

// NewInMemorySummaryLogger creates an in-memory summary logger.
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

type InMemorySummaryLogger struct {
	baseSummaryLogger
}

func (s *InMemorySummaryLogger) GetSummary() string {
	if l, ok := s.baseSummaryLogger.Loggers.(*baselogs.StringLoggers); ok {
		return l.GetLogContent()
	}
	return ""
}
