/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// JSONLoggers defines a JSON logger
type JSONLoggers struct {
	Loggers
	mu           sync.RWMutex
	source       string
	loggerSource string
	writer       WriterWithSource
	zerologger   zerolog.Logger
	closerStore  *parallelisation.CloserStore
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

// Check checks whether the logger is correctly defined or not.
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
	l.zerologger = l.zerologger.With().Timestamp().Logger()
	return nil
}

// Log logs to the output stream.
func (l *JSONLoggers) Log(output ...interface{}) {
	if len(output) == 1 && output[0] == "\n" {
		return
	}
	l.zerologger.Info().Str("source", l.GetLoggerSource()).Msg(fmt.Sprint(output...))
}

// LogError logs to the error stream.
func (l *JSONLoggers) LogError(err ...interface{}) {
	if len(err) == 1 && err[0] == "\n" {
		return
	}
	l.zerologger.Error().Str("source", l.GetLoggerSource()).Msg(fmt.Sprint(err...))
}

// Close closes the logger
func (l *JSONLoggers) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.closerStore.Close()
}

// NewJSONLogger creates a Json logger.
func NewJSONLogger(writer WriterWithSource, loggerSource string, source string) (Loggers, error) {
	return newJSONLogger(true, writer, loggerSource, source)
}

// NewJSONLogger creates a Json logger. It is similar to NewJSONLogger but does not close the writer on Close().
func NewJSONLoggerWithWriter(writer WriterWithSource, loggerSource string, source string) (Loggers, error) {
	return newJSONLogger(false, writer, loggerSource, source)
}

func newJSONLogger(closeWriterOnClose bool, writer WriterWithSource, loggerSource string, source string) (loggers Loggers, err error) {
	closeStore := parallelisation.NewCloserStore(false)
	if closeWriterOnClose {
		closeStore.RegisterCloser(writer)
	}

	zerroLogger := JSONLoggers{
		source:       source,
		loggerSource: loggerSource,
		writer:       writer,
		zerologger:   zerolog.New(writer),
		closerStore:  closeStore,
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

// NewJSONLoggerForSlowWriter creates a lock free, non-blocking & thread safe logger
// wrapped around slowWriter. The slowWriter is closed on Close.
//
// params:
// slowWriter : writer used to write data streams
// ringBufferSize : size of ring buffer used to receive messages
// pollInterval : polling duration to check buffer content
// loggerSource : logger application name
// source : source string
// droppedMessagesLogger : logger for dropped messages
// If pollInterval is greater than 0, a poller is used otherwise a waiter is used.
func NewJSONLoggerForSlowWriter(slowWriter WriterWithSource, ringBufferSize int, pollInterval time.Duration, loggerSource string, source string, droppedMessagesLogger Loggers) (loggers Loggers, err error) {
	return NewJSONLogger(NewDiodeWriterForSlowWriter(slowWriter, ringBufferSize, pollInterval, droppedMessagesLogger), loggerSource, source)
}

// NewJSONLoggerForSlowWriters creates a lock free, non-blocking & thread safe logger
// wrapped around slowWriter. It is similar to NewJSONLoggerForSlowWriter but does not close the writer on Close().
//
// params:
// slowWriter : writer used to write data streams
// ringBufferSize : size of ring buffer used to receive messages
// pollInterval : polling duration to check buffer content
// loggerSource : logger application name
// source : source string
// droppedMessagesLogger : logger for dropped messages
// If pollInterval is greater than 0, a poller is used otherwise a waiter is used.
func NewJSONLoggerForSlowWriters(slowWriter WriterWithSource, ringBufferSize int, pollInterval time.Duration, loggerSource string, source string, droppedMessagesLogger Loggers) (loggers Loggers, err error) {
	return NewJSONLogger(NewDiodeWriter(slowWriter, ringBufferSize, pollInterval, droppedMessagesLogger), loggerSource, source)
}
