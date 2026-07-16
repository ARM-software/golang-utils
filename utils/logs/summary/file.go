package summary

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	baselogs "github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/platform"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

type flushFunc func(string) error
type closeFunc func() error
type clearFunc func() error

// SummaryLogger accumulates summary content in memory.
//
// Summary content is written as plain strings. Destinations may render those
// strings as Markdown or another rich-text format.
type SummaryLogger struct {
	mu           sync.RWMutex
	content      *baselogs.StringLoggers
	flush        flushFunc
	close        closeFunc
	clear        clearFunc
	loggerSource string
	logSource    string
}

// FileSummaryLogger accumulates summary content in memory and flushes it to a
// file-backed destination.
type FileSummaryLogger struct {
	*SummaryLogger
	path       string
	fileLogger baselogs.Loggers
}

// NewSummaryLogger creates an in-memory summary logger.
func NewSummaryLogger(loggerSource string) (logger *SummaryLogger, err error) {
	content, err := baselogs.NewPlainStringLogger()
	if err != nil {
		return
	}
	logger = &SummaryLogger{content: content, loggerSource: loggerSource}
	return
}

// NewFileSummaryLogger creates a summary logger that writes its flushed content
// to path.
func NewFileSummaryLogger(path string, loggerSource string) (logger *FileSummaryLogger, err error) {
	if reflection.IsEmpty(path) {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing summary path")
		return
	}
	fileLogger, err := baselogs.NewFileOnlyLogger(path, loggerSource)
	if err != nil {
		return
	}
	base, err := NewSummaryLogger(loggerSource)
	if err != nil {
		return
	}
	logger = &FileSummaryLogger{
		SummaryLogger: base,
		path:          path,
		fileLogger:    fileLogger,
	}
	logger.SummaryLogger.flush = logger.flushToFile
	logger.SummaryLogger.close = fileLogger.Close
	logger.SummaryLogger.clear = logger.clearFile
	return
}

func (l *SummaryLogger) Check() error {
	if l == nil || l.content == nil {
		return commonerrors.ErrNoLogger
	}
	return nil
}

func (l *SummaryLogger) Close() (err error) {
	if err = l.Check(); err != nil {
		return
	}
	if l.flush != nil {
		err = l.flush(l.Content())
		if err != nil {
			return
		}
	}
	if l.close != nil {
		err = l.close()
	}
	return
}

func (l *SummaryLogger) SetLogSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLogSource
	}
	l.mu.Lock()
	l.logSource = source
	l.mu.Unlock()
	return nil
}

func (l *SummaryLogger) SetLoggerSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLoggerSource
	}
	l.mu.Lock()
	l.loggerSource = source
	l.mu.Unlock()
	return nil
}

func (l *SummaryLogger) Log(output ...interface{}) {
	_ = l.WriteLine(l.formatLine(fmt.Sprint(output...)))
}

func (l *SummaryLogger) LogError(err ...interface{}) {
	_ = l.WriteLine(l.formatLine(fmt.Sprintf("ERROR: %s", fmt.Sprint(err...))))
}

func (l *SummaryLogger) WriteString(content string) (err error) {
	if err = l.Check(); err != nil {
		return
	}
	_, err = l.content.LogWriter.Write([]byte(content))
	return
}

func (l *SummaryLogger) WriteStringF(format string, values ...any) error {
	return l.WriteString(fmt.Sprintf(format, values...))
}

func (l *SummaryLogger) WriteLine(content string) error {
	if strings.HasSuffix(content, platform.LineSeparator()) {
		return l.WriteString(content)
	}
	return l.WriteString(content + platform.LineSeparator())
}

func (l *SummaryLogger) WriteLineF(format string, values ...any) error {
	return l.WriteLine(fmt.Sprintf(format, values...))
}

func (l *SummaryLogger) Flush() (err error) {
	if err = l.Check(); err != nil {
		return
	}
	if l.flush != nil {
		err = l.flush(l.Content())
	}
	return
}

func (l *SummaryLogger) Content() string {
	if l == nil || l.content == nil {
		return ""
	}
	return l.content.GetLogContent()
}

func (l *SummaryLogger) Clear() (err error) {
	if err = l.Check(); err != nil {
		return
	}
	err = l.content.LogWriter.Close()
	if err != nil {
		return
	}
	if l.clear != nil {
		err = l.clear()
	}
	return
}

func (l *SummaryLogger) formatLine(message string) string {
	l.mu.RLock()
	loggerSource := l.loggerSource
	logSource := l.logSource
	l.mu.RUnlock()

	prefix := ""
	switch {
	case reflection.IsNotEmpty(logSource):
		prefix = fmt.Sprintf("[%s] ", logSource)
	case reflection.IsNotEmpty(loggerSource):
		prefix = fmt.Sprintf("[%s] ", loggerSource)
	}
	return strings.TrimSpace(prefix + message)
}

var _ ISummaryLogger = (*SummaryLogger)(nil)
var _ ISummaryLogger = (*FileSummaryLogger)(nil)

func (l *FileSummaryLogger) flushToFile(content string) error {
	if reflection.IsEmpty(l.path) {
		return commonerrors.New(commonerrors.ErrUndefined, "missing summary path")
	}
	file, err := filesystem.OpenFile(l.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	_ = file.Close()
	if reflection.IsEmpty(content) {
		return nil
	}
	return filesystem.WriteFile(l.path, []byte(content), 0o644)
}

func (l *FileSummaryLogger) clearFile() error {
	if reflection.IsEmpty(l.path) {
		return commonerrors.New(commonerrors.ErrUndefined, "missing summary path")
	}
	file, err := filesystem.OpenFile(l.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}
