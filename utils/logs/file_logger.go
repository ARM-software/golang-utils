/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import "github.com/sirupsen/logrus"

// NewFileLogger creates a logger to a file
func NewFileLogger(logFile string, loggerSource string) (loggers Loggers, err error) {
	return NewLogrusLoggerWithFileHook(logrus.New(), loggerSource, logFile)

}

// CreateFileLogger creates a logger to a file
//
// Deprecated: Use NewFileLogger instead
func CreateFileLogger(logFile string, loggerSource string) (loggers Loggers, err error) {
	return NewFileLogger(logFile, loggerSource)
}
