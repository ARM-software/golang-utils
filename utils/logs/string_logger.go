/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

type StringWriter struct {
	io.WriteCloser
	mu   sync.RWMutex
	Logs strings.Builder
}

func (w *StringWriter) Write(p []byte) (n int, err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.Logs.Write(p)
	return
}

func (w *StringWriter) Close() (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Logs.Reset()
	return
}

func (w *StringWriter) GetFullContent() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Logs.String()
}

type StringLoggers struct {
	GenericLoggers
	LogWriter StringWriter
}

func (l *StringLoggers) Check() error {
	return l.GenericLoggers.Check()
}

func (l *StringLoggers) GetLogContent() string {
	return l.LogWriter.GetFullContent()
}

// Close closes the logger
func (l *StringLoggers) Close() (err error) {
	err = l.LogWriter.Close()
	if err != nil {
		return
	}
	err = l.GenericLoggers.Close()
	return
}

// NewStringLogger creates a logger to standard output/error
func NewStringLogger(loggerSource string) (loggers *StringLoggers, err error) {
	loggers = &StringLoggers{
		LogWriter: StringWriter{},
	}
	loggers.GenericLoggers = GenericLoggers{
		Output: log.New(&loggers.LogWriter, fmt.Sprintf("[%v] Output: ", loggerSource), log.LstdFlags),
		Error:  log.New(&loggers.LogWriter, fmt.Sprintf("[%v] Error: ", loggerSource), log.LstdFlags),
	}
	return
}

// CreateStringLogger creates a logger to standard output/error
//
// Deprecated: Use NewStringLogger instead
func CreateStringLogger(loggerSource string) (loggers *StringLoggers, err error) {
	return NewStringLogger(loggerSource)
}
