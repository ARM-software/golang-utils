package summary

import (
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// NewFileSummaryLogger creates a summary logger that writes summary entries to
// path as plain text.
//
// This is useful in CI and reporting flows where the final summary should be
// written to a known file for later display or collection.
func NewFileSummaryLogger(path string, loggerSource string) (logger *FileSummaryLogger, err error) {
	if reflection.IsEmpty(path) {
		err = commonerrors.UndefinedVariable("log file path")
		return
	}
	bLogger, err := baselogs.NewTextFileOnlyLogger(path, loggerSource)
	if err != nil {
		return
	}
	logger = &FileSummaryLogger{NewSummaryLogger(bLogger)}
	return
}

// FileSummaryLogger writes summary content to a plain-text file sink.
//
// The file content may still contain Markdown syntax; the logger itself only
// treats it as plain text.
type FileSummaryLogger struct {
	ISummaryLogger
}
