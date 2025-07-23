/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safecast"
	sizeUnits "github.com/ARM-software/golang-utils/utils/units/size"
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

// NewRollingFilesLogger creates a rolling file logger using [lumberjack](https://github.com/natefinch/lumberjack) under the bonnet.
// maxSize is the maximum size in bytes of a log file before it gets rotated.
// maxBackups is the maximum number of old log files to retain.
// maxAge is the maximum duration old log files are retained.
func NewRollingFilesLogger(logFile string, loggerSource string, maxFileSize float64, maxBackups int, maxAge time.Duration) (loggers Loggers, err error) {
	if reflection.IsEmpty(logFile) {
		err = commonerrors.New(commonerrors.ErrInvalidDestination, "missing file destination")
		return
	}
	l := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    safecast.ToInt(maxFileSize / sizeUnits.MiB),
		MaxAge:     safecast.ToInt(maxAge.Hours() / 24),
		MaxBackups: maxBackups,
		LocalTime:  false,
		Compress:   false,
	}
	closerStore := parallelisation.NewCloserStore(false)
	closerStore.RegisterCloser(l)

	loggers = &GenericLoggers{
		Output:     log.New(l, fmt.Sprintf("[%v] Output: ", loggerSource), log.LstdFlags),
		Error:      log.New(l, fmt.Sprintf("[%v] Error: ", loggerSource), log.LstdFlags),
		closeStore: closerStore,
	}
	return
}
