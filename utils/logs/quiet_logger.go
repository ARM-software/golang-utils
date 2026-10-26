/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import "github.com/ARM-software/golang-utils/utils/commonerrors"

type quietLogger struct {
	loggers Loggers
}

func (l *quietLogger) Close() error {
	return l.loggers.Close()
}

func (l *quietLogger) Check() error {
	return l.loggers.Check()
}

func (l *quietLogger) SetLogSource(source string) error {
	return l.loggers.SetLogSource(source)
}

func (l *quietLogger) SetLoggerSource(source string) error {
	return l.loggers.SetLoggerSource(source)
}

func (l *quietLogger) Log(_ ...interface{}) {
}

func (l *quietLogger) LogError(err ...interface{}) {
	l.loggers.LogError(err...)
}

// NewQuietLogger returns a quiet logger which only logs errors.
func NewQuietLogger(loggers Loggers) (Loggers, error) {
	if loggers == nil {
		return nil, commonerrors.ErrNoLogger
	}
	return &quietLogger{loggers: loggers}, nil
}
