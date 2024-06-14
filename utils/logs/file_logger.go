/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"io"

	"github.com/sirupsen/logrus"
)

// NewFileLogger creates a logger to a file.
func NewFileLogger(logFile string, loggerSource string) (loggers Loggers, err error) {
	return NewLogrusLoggerWithFileHook(logrus.New(), loggerSource, logFile)
}

// NewFileOnlyLogger creates a logger to a file such as NewFileLogger but logs are only sent to a file and will not be printed to StdErr or StdOut.
func NewFileOnlyLogger(logFile string, loggerSource string) (loggers Loggers, err error) {
	underlying := logrus.New()
	underlying.SetOutput(io.Discard)
	return NewLogrusLoggerWithFileHook(underlying, loggerSource, logFile)
}

// CreateFileLogger creates a logger to a file
//
// Deprecated: Use NewFileLogger instead
func CreateFileLogger(logFile string, loggerSource string) (loggers Loggers, err error) {
	return NewFileLogger(logFile, loggerSource)
}
