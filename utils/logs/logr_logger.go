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
	logger logr.Logger
}

func (l *logrLogger) Close() error {
	return nil
}

func (l *logrLogger) Check() error {
	if l.logger == nil {
		return commonerrors.ErrNoLogger
	}
	if l.logger.Enabled() {
		return nil
	}
	return fmt.Errorf("%w: disabled logger", commonerrors.ErrCondition)
}

func (l *logrLogger) SetLogSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLogSource
	}
	l.logger.WithValues(KeyLogSource, source)
	return nil
}

func (l *logrLogger) SetLoggerSource(source string) error {
	if reflection.IsEmpty(source) {
		return commonerrors.ErrNoLoggerSource
	}
	l.logger.WithName(source)
	l.logger.WithValues(KeyLoggerSource, source)
	return nil
}

func (l *logrLogger) Log(output ...interface{}) {
	l.logger.Info(fmt.Sprintln(output...))
}

func (l *logrLogger) LogError(err ...interface{}) {
	l.logger.Error(nil, fmt.Sprintln(err...))
}

// NewLogrLogger creates loggers based on a logr implementation (https://github.com/go-logr/logr)
func NewLogrLogger(logrImpl logr.Logger, loggerSource string) (loggers Loggers, err error) {
	loggers = &logrLogger{logger: logrImpl}
	err = loggers.SetLoggerSource(loggerSource)
	return
}

// NewLogrLoggerFromLoggers converts loggers into a logr.Logger
func NewLogrLoggerFromLoggers(loggers Loggers) logr.Logger {
	return stdr.New(newGolangStdLoggerFromLoggers(loggers))
}
