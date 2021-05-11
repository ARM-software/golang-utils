package logs

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ARMmbed/golang-utils/utils/commonerrors"
)

type Loggers interface {
	io.Closer
	// Checks whether the loggers are correctly defined or not.
	Check() error
	// Sets the source of the log message e.g. related build job, related command, etc.
	SetLogSource(source string) error
	// Sets the source of the logger e.g. APIs, Build worker, CMSIS tools.
	SetLoggerSource(source string) error
	// Logs to the output logger.
	Log(output ...interface{})
	// Logs to the Error logger.
	LogError(err ...interface{})
}

// Definition of command loggers
type GenericLoggers struct {
	Output *log.Logger
	Error  *log.Logger
}

// Checks whether the loggers are correctly defined or not.
func (l *GenericLoggers) Check() error {
	if l.Error == nil || l.Output == nil {
		return commonerrors.ErrNoLogger
	}
	return nil
}

func (l *GenericLoggers) SetLogSource(source string) error {
	return nil
}

func (l *GenericLoggers) SetLoggerSource(source string) error {
	return nil
}

// Logs to the output logger.
func (l *GenericLoggers) Log(output ...interface{}) {
	l.Output.Println(output...)
}

// Logs to the Error logger.
func (l *GenericLoggers) LogError(err ...interface{}) {
	l.Error.Println(err...)
}

// Closes the logger
func (l *GenericLoggers) Close() error {
	return nil
}

type AsynchronousLoggers struct {
	mu           sync.RWMutex
	oWriter      WriterWithSource
	eWriter      WriterWithSource
	loggerSource string
}

func (l *AsynchronousLoggers) Check() error {
	if l.GetLoggerSource() == "" {
		return commonerrors.ErrNoLoggerSource
	}
	if l.oWriter == nil || l.eWriter == nil {
		return commonerrors.ErrUndefined
	}
	return nil
}

func (l *AsynchronousLoggers) SetLogSource(source string) error {
	err1 := l.oWriter.SetSource(source)
	err2 := l.eWriter.SetSource(source)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (l *AsynchronousLoggers) SetLoggerSource(source string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loggerSource = source
	return nil
}

func (l *AsynchronousLoggers) GetLoggerSource() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loggerSource
}

func (l *AsynchronousLoggers) Log(output ...interface{}) {
	_, _ = l.oWriter.Write([]byte(fmt.Sprintf("[%v] Output (%v): %v\n", l.GetLoggerSource(), time.Now(), strings.TrimSpace(fmt.Sprint(output...)))))
}

func (l *AsynchronousLoggers) LogError(err ...interface{}) {
	_, _ = l.eWriter.Write([]byte(fmt.Sprintf("[%v] Error (%v): %v\n", l.GetLoggerSource(), time.Now(), strings.TrimSpace(fmt.Sprint(err...)))))
}

func (l *AsynchronousLoggers) Close() error {
	err1 := l.eWriter.Close()
	err2 := l.oWriter.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func NewAsynchronousLoggers(slowOutputWriter WriterWithSource, slowErrorWriter WriterWithSource, ringBufferSize int, pollInterval time.Duration, loggerSource string, source string, droppedMessagesLogger Loggers) (loggers Loggers, err error) {
	loggers = &AsynchronousLoggers{
		oWriter:      NewDiodeWriterForSlowWriter(slowOutputWriter, ringBufferSize, pollInterval, droppedMessagesLogger),
		eWriter:      NewDiodeWriterForSlowWriter(slowErrorWriter, ringBufferSize, pollInterval, droppedMessagesLogger),
		loggerSource: loggerSource,
	}
	err = loggers.SetLogSource(source)
	if err != nil {
		return
	}
	err = loggers.Check()
	return
}
