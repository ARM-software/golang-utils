/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import "io"

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE Loggers,WriterWithSource

type Loggers interface {
	io.Closer
	// Checks whether the loggers are correctly defined or not.
	Check() error
	// Sets the source of the log message e.g. related build job, related command, etc.
	SetLogSource(source string) error
	// Sets the source of the logger e.g. APIs, Build worker, CMSIS tools.
	SetLoggerSource(source string) error
	// Logs to the output logger.
	Log(output ...interface{})
	// Logs to the Error logger.
	LogError(err ...interface{})
}

type WriterWithSource interface {
	io.WriteCloser
	SetSource(source string) error
}
