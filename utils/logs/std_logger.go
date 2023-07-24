/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ARM-software/golang-utils/utils/platform"
)

type StdWriter struct {
	WriterWithSource
}

func (w *StdWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (w *StdWriter) Close() error {
	return nil
}

func (w *StdWriter) SetSource(source string) error {
	_, err := os.Stdout.Write([]byte(fmt.Sprintf("Source: %v%v", source, platform.LineSeparator())))
	return err
}

type StdErrWriter struct {
	WriterWithSource
}

func (w *StdErrWriter) Write(p []byte) (n int, err error) {
	return os.Stderr.Write(p)
}

func (w *StdErrWriter) Close() error {
	return nil
}

func (w *StdErrWriter) SetSource(source string) error {
	return nil
}

// NewStdLogger creates a logger to standard output/error
func NewStdLogger(loggerSource string) (loggers Loggers, err error) {
	loggers = &GenericLoggers{
		Output: log.New(os.Stdout, fmt.Sprintf("[%v] Output: ", loggerSource), log.LstdFlags),
		Error:  log.New(os.Stderr, fmt.Sprintf("[%v] Error: ", loggerSource), log.LstdFlags),
	}
	return
}

// CreateStdLogger creates a logger to standard output/error
//
// Deprecated: use NewStdLogger instead
func CreateStdLogger(loggerSource string) (loggers Loggers, err error) {
	return NewStdLogger(loggerSource)
}

func NewAsynchronousStdLogger(loggerSource string, ringBufferSize int, pollInterval time.Duration, source string) (loggers Loggers, err error) {
	return NewAsynchronousLoggers(&StdWriter{}, &StdErrWriter{}, ringBufferSize, pollInterval, loggerSource, source, nil)
}

func newGolangStdLoggerFromLoggers(loggers Loggers) StdLogger {
	return &stdAdaptor{
		loggers: loggers,
	}
}

type stdAdaptor struct {
	loggers Loggers
}

func (s *stdAdaptor) Output(_ int, logline string) error {
	s.loggers.Log(logline)
	return nil
}
