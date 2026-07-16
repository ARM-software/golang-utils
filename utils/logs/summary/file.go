package summary

import (
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// NewFileSummaryLogger creates a summary logger that writes appended summary
// content to path.
func NewFileSummaryLogger(path string, loggerSource string) (logger *FileSummaryLogger, err error) {
	if reflection.IsEmpty(path) {
		err = commonerrors.UndefinedVariable("log file path")
		return
	}
	bLogger, err := baselogs.NewTextFileOnlyLogger(path, loggerSource)
	if err != nil {
		return
	}
	logger = &FileSummaryLogger{baseSummaryLogger{
		Loggers: bLogger,
	}}
	return
}

// FileSummaryLogger writes summary content to a plain-text file sink.
type FileSummaryLogger struct {
	baseSummaryLogger
}
