/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package logs defines loggers for use in projects.
package logs

import (
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs/logrimp"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// NewLogrusLogger returns a logger which uses logrus logger (https://github.com/Sirupsen/logrus)
func NewLogrusLogger(logrusL *logrus.Logger, loggerSource string) (loggers Loggers, err error) {
	if logrusL == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	return NewLogrLogger(logrimp.NewLogrusLogger(logrusL), loggerSource)
}

// NewLogrusLoggerWithFileHook returns a logger which uses a logrus logger (https://github.com/Sirupsen/logrus) and writes the logs to `logFilePath`
func NewLogrusLoggerWithFileHook(logrusL *logrus.Logger, loggerSource string, logFilePath string) (loggers Loggers, err error) {
	if logrusL == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	if reflection.IsEmpty(logFilePath) {
		err = commonerrors.New(commonerrors.ErrInvalidDestination, "missing file destination")
		return
	}
	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  logFilePath,
		logrus.ErrorLevel: logFilePath,
	}
	logrusL.Hooks.Add(lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{},
	))
	return NewLogrusLogger(logrusL, loggerSource)
}
