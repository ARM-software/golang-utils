package logs

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/ARMmbed/golang-utils/utils/commonerrors"
)

// Definition of JSON message loggers
type JSONLoggers struct {
	Loggers
	mu           sync.RWMutex
	source       string
	loggerSource string
	writer       WriterWithSource
	Zerologger   zerolog.Logger
}

func (l *JSONLoggers) SetLogSource(source string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.source = source
	return l.writer.SetSource(source)
}

func (l *JSONLoggers) SetLoggerSource(source string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loggerSource = source
	return nil
}

func (l *JSONLoggers) GetSource() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.source
}
func (l *JSONLoggers) GetLoggerSource() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loggerSource
}

// Checks whether the loggers are correctly defined or not.
func (l *JSONLoggers) Check() error {
	if l.GetSource() == "" {
		return commonerrors.ErrNoLogSource
	}
	if l.GetLoggerSource() == "" {
		return commonerrors.ErrNoLoggerSource
	}
	return nil
}

func (l *JSONLoggers) Configure() error {
	zerolog.TimestampFieldName = "ctime"
	zerolog.MessageFieldName = "message"
	zerolog.LevelFieldName = "severity"
	l.Zerologger = l.Zerologger.With().Timestamp().Logger()
	return nil
}

// Logs to the output logger.
func (l *JSONLoggers) Log(output ...interface{}) {
	if len(output) == 1 && output[0] == "\n" {
		return
	}
	l.Zerologger.Info().Str("source", l.GetLoggerSource()).Msg(fmt.Sprint(output...))
}

// Logs to the Error logger.
func (l *JSONLoggers) LogError(err ...interface{}) {
	if len(err) == 1 && err[0] == "\n" {
		return
	}
	l.Zerologger.Error().Str("source", l.GetLoggerSource()).Msg(fmt.Sprint(err...))
}

// Closes the logger
func (l *JSONLoggers) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.writer.Close()
}

func NewJSONLogger(writer WriterWithSource, loggerSource string, source string) (loggers Loggers, err error) {
	zerroLogger := JSONLoggers{
		source:       source,
		loggerSource: loggerSource,
		writer:       writer,
		Zerologger:   zerolog.New(writer),
	}
	err = zerroLogger.Check()
	if err != nil {
		return
	}
	err = writer.SetSource(source)
	if err != nil {
		return
	}
	err = zerroLogger.Configure()
	loggers = &zerroLogger
	return
}

// NewJSONLoggerForSlowWriter creates a lock free, non blocking & thread safe logger
// wrapped around slowWriter
//
// params:
//		slowWriter : writer used to write data streams
// 		ringBufferSize : size of ring buffer used to receive messages
// 		pollInterval : polling duration to check buffer content
//		loggerSource : logger application name
//		source : source string
//		droppedMessagesLogger : logger for dropped messages
//
// If pollInterval is greater than 0, a poller is used otherwise a waiter is used.
func NewJSONLoggerForSlowWriter(slowWriter WriterWithSource, ringBufferSize int, pollInterval time.Duration, loggerSource string, source string, droppedMessagesLogger Loggers) (loggers Loggers, err error) {
	return NewJSONLogger(NewDiodeWriterForSlowWriter(slowWriter, ringBufferSize, pollInterval, droppedMessagesLogger), loggerSource, source)
}
