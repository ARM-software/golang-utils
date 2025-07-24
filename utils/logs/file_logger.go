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

	"github.com/DeRuina/timberjack"
	"github.com/sirupsen/logrus"

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

type FileLoggerOptions struct {
	maxFileSize float64
	maxAge      time.Duration
	maxBackups  int
}

type FileLoggerOption func(*FileLoggerOptions) *FileLoggerOptions

// WithMaxFileSize sets the maximum size in bytes of a log file before it gets rotated.
func WithMaxFileSize(maxFileSize float64) FileLoggerOption {
	return func(o *FileLoggerOptions) *FileLoggerOptions {
		if o == nil {
			return o
		}
		o.maxFileSize = maxFileSize
		return o
	}
}

// WithMaxAge sets the maximum duration old log files are retained.
func WithMaxAge(maxAge time.Duration) FileLoggerOption {
	return func(o *FileLoggerOptions) *FileLoggerOptions {
		if o == nil {
			return o
		}
		// This is necessary to avoid a Race Condition
		if maxAge >= time.Minute {
			o.maxAge = maxAge
		}
		return o
	}
}

// WithMaxBackups sets the maximum number of old log files to retain.
func WithMaxBackups(maxBackups int) FileLoggerOption {
	return func(o *FileLoggerOptions) *FileLoggerOptions {
		if o == nil {
			return o
		}
		o.maxBackups = maxBackups
		return o
	}
}

// NewRollingFilesLogger creates a rolling file logger using [lumberjack](https://github.com/natefinch/lumberjack) under the bonnet.
func NewRollingFilesLogger(logFile string, loggerSource string, options ...FileLoggerOption) (loggers Loggers, err error) {
	opts := &FileLoggerOptions{
		maxFileSize: 100 * sizeUnits.MiB,
		maxAge:      24 * time.Hour,
		maxBackups:  3,
	}
	for i := range options {
		opts = options[i](opts)
	}
	if reflection.IsEmpty(logFile) {
		err = commonerrors.New(commonerrors.ErrInvalidDestination, "missing file destination")
		return
	}
	l := &timberjack.Logger{
		Filename:   logFile,
		MaxSize:    safecast.ToInt(opts.maxFileSize / sizeUnits.MiB),
		MaxAge:     safecast.ToInt(opts.maxAge.Hours() / 24),
		MaxBackups: opts.maxBackups,
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
