package summary

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

type summaryLogger struct {
	mu           sync.RWMutex
	writer       Writer
	loggerSource string
	logSource    string
}

// NewLogger creates a summary logger backed by writer.
func NewLogger(writer Writer, loggerSource string) (Loggers, error) {
	logger := &summaryLogger{writer: writer, loggerSource: loggerSource}
	if err := logger.Check(); err != nil {
		return nil, err
	}
	return logger, nil
}

func (l *summaryLogger) Close() error {
	if l.writer == nil {
		return nil
	}
	return l.writer.Close()
}

func (l *summaryLogger) Check() error {
	if l.writer == nil {
		return commonerrors.ErrNoLogger
	}
	return nil
}

func (l *summaryLogger) SetLogSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLogSource
	}
	l.mu.Lock()
	l.logSource = source
	l.mu.Unlock()
	return nil
}

func (l *summaryLogger) SetLoggerSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLoggerSource
	}
	l.mu.Lock()
	l.loggerSource = source
	l.mu.Unlock()
	return nil
}

func (l *summaryLogger) Log(output ...interface{}) {
	_ = l.WriteMarkdown(l.formatLine(fmt.Sprint(output...)))
}

func (l *summaryLogger) LogError(err ...interface{}) {
	message := fmt.Sprint(err...)
	_ = l.WriteMarkdown(l.formatLine(fmt.Sprintf("ERROR: %s", message)))
}

func (l *summaryLogger) WriteMarkdown(markdown string) error {
	if err := l.Check(); err != nil {
		return err
	}
	return l.writer.WriteMarkdown(markdown)
}

func (l *summaryLogger) Clear() error {
	if err := l.Check(); err != nil {
		return err
	}
	return l.writer.Clear()
}

func (l *summaryLogger) formatLine(message string) string {
	l.mu.RLock()
	loggerSource := l.loggerSource
	logSource := l.logSource
	l.mu.RUnlock()

	prefix := ""
	switch {
	case !reflection.IsEmpty(logSource):
		prefix = fmt.Sprintf("[%s] ", logSource)
	case !reflection.IsEmpty(loggerSource):
		prefix = fmt.Sprintf("[%s] ", loggerSource)
	}
	return fmt.Sprintf("%s%s\n", prefix, strings.TrimSpace(message))
}

var _ baselogs.Loggers = (*summaryLogger)(nil)
