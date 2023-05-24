/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
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
	if l.logger.IsZero() {
		return commonerrors.ErrNoLogger
	}
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
	return stdr.New(newGolangStdLoggerFromLoggers(loggers))
}
