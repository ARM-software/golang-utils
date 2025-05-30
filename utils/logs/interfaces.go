/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package logs defines loggers for use in projects.
package logs

import (
	"io"

	"github.com/go-logr/logr"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE Loggers,IMultipleLoggers,WriterWithSource,StdLogger

// Loggers defines generic loggers which separate common logging messages from errors.
// This is to use in cases where it is necessary to separate the two streams e.g. remote procedure call (RPC)
// In most cases however, if only a standard logger is needed, it is advised to use logr.Logger.
type Loggers interface {
	io.Closer
	// Check returns whether the loggers are correctly defined or not.
	Check() error
	// SetLogSource sets the source of the log message e.g. related build job, related command, etc.
	SetLogSource(source string) error
	// SetLoggerSource sets the source of the logger e.g. APIs, Build worker, CMSIS tools.
	SetLoggerSource(source string) error
	// Log logs to the output stream/logger.
	Log(output ...interface{})
	// LogError logs to the Error stream/logger.
	LogError(err ...interface{})
}

// IMultipleLoggers provides an interface to manage multiple loggers the same way as a single logger.
type IMultipleLoggers interface {
	Loggers
	// AppendLogger appends generic loggers to the internal list of loggers managed by this system.
	AppendLogger(l ...logr.Logger) error
	// Append appends loggers to the internal list of loggers managed by this system.
	Append(l ...Loggers) error
}

type WriterWithSource interface {
	io.WriteCloser
	SetSource(source string) error
}

// StdLogger is the subset of the Go stdlib log.Logger API.
type StdLogger interface {
	// Output is the same as log.Output and log.Logger.Output.
	Output(calldepth int, logline string) error
}
