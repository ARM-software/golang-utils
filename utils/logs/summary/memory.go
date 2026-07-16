package summary

// NewInMemorySummaryLogger creates an in-memory summary logger backed by the
// repository's plain string logger implementation.
func NewInMemorySummaryLogger(loggerSource string) (logger *SummaryLogger, err error) {
	logger, err = NewSummaryLogger(loggerSource)
	return
}
