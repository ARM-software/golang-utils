package logs

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	"github.com/ARMmbed/golang-utils/utils/commonerrors"
)

// Definition of JSON message loggers
type JsonLoggers struct {
	Loggers
	mu           sync.RWMutex
	source       string
	loggerSource string
	writer       WriterWithSource
	Zerologger   zerolog.Logger
}

func (l *JsonLoggers) SetLogSource(source string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.source = source
	return l.writer.SetSource(source)
}

func (l *JsonLoggers) SetLoggerSource(source string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loggerSource = source
	return nil
}

func (l *JsonLoggers) GetSource() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.source
}
func (l *JsonLoggers) GetLoggerSource() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loggerSource
}

// Checks whether the loggers are correctly defined or not.
func (l *JsonLoggers) Check() error {
	if l.GetSource() == "" {
		return commonerrors.ErrNoLogSource
	}
	if l.GetLoggerSource() == "" {
		return commonerrors.ErrNoLoggerSource
	}
	return nil
}

func (l *JsonLoggers) Configure() error {
	zerolog.TimestampFieldName = "ctime"
	zerolog.MessageFieldName = "message"
	zerolog.LevelFieldName = "severity"
	l.Zerologger = l.Zerologger.With().Timestamp().Logger()
	return nil
}

// Logs to the output logger.
func (l *JsonLoggers) Log(output ...interface{}) {
	if len(output) == 1 && output[0] == "\n" {
		return
	}
	l.Zerologger.Info().Str("source", l.GetLoggerSource()).Msg(fmt.Sprint(output...))
}

// Logs to the Error logger.
func (l *JsonLoggers) LogError(err ...interface{}) {
	if len(err) == 1 && err[0] == "\n" {
		return
	}
	l.Zerologger.Error().Str("source", l.GetLoggerSource()).Msg(fmt.Sprint(err...))
}

// Closes the logger
func (l *JsonLoggers) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.writer.Close()
}

func CreateJsonLogger(writer WriterWithSource, loggerSource string, source string) (loggers Loggers, err error) {
	zerroLogger := JsonLoggers{
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

func CreateJsonLoggerForSlowWriter(slowWriter WriterWithSource, loggerSource string, source string, droppedMessagesLogger Loggers) (loggers Loggers, err error) {
	return CreateJsonLogger(NewDiodeWriterForSlowWriter(slowWriter, droppedMessagesLogger), loggerSource, source)
}
