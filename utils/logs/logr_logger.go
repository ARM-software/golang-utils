/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	KeyLogSource    = "source"
	KeyLoggerSource = "logger-source"
)

type logrLogger struct {
	logger    logr.Logger
	closeFunc func() error
}

func (l *logrLogger) Close() error {
	if l.closeFunc != nil {
		return l.closeFunc()
	}
	return nil
}

func (l *logrLogger) Check() error {
	return nil
}

func (l *logrLogger) SetLogSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLogSource
	}
	l.logger = l.logger.WithValues(KeyLogSource, source)
	return nil
}

func (l *logrLogger) SetLoggerSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLoggerSource
	}
	l.logger = l.logger.WithName(source).WithValues(KeyLoggerSource, source)
	return nil
}

func (l *logrLogger) Log(output ...interface{}) {
	l.logger.Info(fmt.Sprintln(output...))
}

func (l *logrLogger) LogError(err ...interface{}) {
	if len(err) > 0 {
		if subErr, ok := err[0].(error); ok {
			l.logger.Error(subErr, fmt.Sprintln(err...))
		} else {
			l.logger.Error(nil, fmt.Sprintln(err...))
		}
	} else {
		l.logger.Error(nil, "")
	}

}

// NewLogrLogger creates loggers based on a logr implementation (https://github.com/go-logr/logr)
func NewLogrLogger(logrImpl logr.Logger, loggerSource string) (Loggers, error) {
	return NewLogrLoggerWithClose(logrImpl, loggerSource, nil)
}

// NewLogrLoggerWithClose creates loggers based on a logr implementation (https://github.com/go-logr/logr)
func NewLogrLoggerWithClose(logrImpl logr.Logger, loggerSource string, closeFunc func() error) (loggers Loggers, err error) {
	loggers = &logrLogger{logger: logrImpl, closeFunc: closeFunc}
	err = loggers.SetLoggerSource(loggerSource)
	return
}

// NewLogrLoggerFromLoggers converts loggers into a logr.Logger
func NewLogrLoggerFromLoggers(loggers Loggers) logr.Logger {
	return logr.New(NewLoggersLogSink(loggers))
}

// NewPlainLogrLoggerFromLoggers converts loggers into a logr.Logger but do not print any data other than the messages
func NewPlainLogrLoggerFromLoggers(loggers Loggers) logr.Logger {
	return logr.New(NewPlainLoggersSink(loggers))
}

// GetLogrLoggerFromContext gets a logger from a context, unless it does not exist then it returns an ErrNoLogger
func GetLogrLoggerFromContext(ctx context.Context) (logger logr.Logger, err error) {
	logger, err = logr.FromContext(ctx)
	if err != nil {
		err = fmt.Errorf("%w: %v", commonerrors.ErrNoLogger, err.Error())
		return
	}
	if logger.IsZero() {
		err = commonerrors.ErrNoLogger
	}
	return
}

type plainLoggersSinkAdapter struct {
	underlying Loggers
}

func (s *plainLoggersSinkAdapter) Init(_ logr.RuntimeInfo) {
}

func (s *plainLoggersSinkAdapter) Enabled(_ int) bool {
	return true
}

func (s *plainLoggersSinkAdapter) Info(level int, msg string, keysAndValues ...any) {
	if s.underlying != nil {
		s.underlying.Log(msg)
	}
}

func (s *plainLoggersSinkAdapter) Error(err error, msg string, keysAndValues ...any) {
	if s.underlying != nil {
		if err == nil {
			s.underlying.LogError(msg)
		} else {
			s.underlying.LogError(fmt.Sprintf("%v: %v", msg, err.Error()))
		}
	}
}

func (s *plainLoggersSinkAdapter) WithValues(keysAndValues ...any) logr.LogSink {
	return &plainLoggersSinkAdapter{underlying: s.underlying}
}

func (s *plainLoggersSinkAdapter) WithName(name string) logr.LogSink {
	if s.underlying != nil {
		_ = s.underlying.SetLogSource(name)
	}
	return &plainLoggersSinkAdapter{underlying: s.underlying}
}

func NewPlainLoggersSink(logger Loggers) logr.LogSink {
	return &plainLoggersSinkAdapter{
		underlying: logger,
	}
}

type loggersLogSinkAdapter struct {
	stdOut     logr.LogSink
	stdErr     logr.LogSink
	underlying Loggers
}

func (l *loggersLogSinkAdapter) Init(info logr.RuntimeInfo) {
	l.stdOut.Init(info)
	l.stdErr.Init(info)
}

func (l *loggersLogSinkAdapter) Enabled(level int) bool {
	return l.stdOut.Enabled(level) &&
		l.stdErr.Enabled(level)
}

func (l *loggersLogSinkAdapter) Info(level int, msg string, keysAndValues ...any) {
	l.stdOut.Info(level, msg, keysAndValues...)
}

func (l *loggersLogSinkAdapter) Error(err error, msg string, keysAndValues ...any) {
	l.stdErr.Error(err, msg, keysAndValues...)
}

func (l *loggersLogSinkAdapter) WithValues(keysAndValues ...any) logr.LogSink {
	return &loggersLogSinkAdapter{
		stdOut:     l.stdOut.WithValues(keysAndValues...),
		stdErr:     l.stdErr.WithValues(keysAndValues...),
		underlying: l.underlying,
	}
}

func (l *loggersLogSinkAdapter) WithName(name string) logr.LogSink {
	if l.underlying != nil {
		_ = l.underlying.SetLogSource(name)
	}
	return &loggersLogSinkAdapter{
		stdOut:     l.stdOut.WithName(name),
		stdErr:     l.stdErr.WithName(name),
		underlying: l.underlying,
	}
}

func NewLoggersLogSink(logger Loggers) logr.LogSink {
	return &loggersLogSinkAdapter{
		stdOut:     stdr.New(NewGolangStdLoggerFromLoggers(logger, false)).GetSink(),
		stdErr:     stdr.New(NewGolangStdLoggerFromLoggers(logger, true)).GetSink(),
		underlying: logger,
	}
}
