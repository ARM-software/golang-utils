/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"io"
	"strings"

	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/platform"
)

var lineSep = platform.UnixLineSeparator()

// INTERNAL
// Way of redirecting process output to a logger.
type logStreamer struct {
	io.Writer
	IsStdErr bool
	Loggers  logs.Loggers
}

func (l *logStreamer) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), lineSep)
	for i := range lines { // https://stackoverflow.com/questions/62446118/implicit-memory-aliasing-in-for-loop
		line := lines[i]
		if line != "" {
			if l.IsStdErr {
				l.Loggers.LogError(line)
			} else {
				l.Loggers.Log(line)
			}
		}
	}
	return len(p), nil
}

func newLogStreamer(isStdErr bool, loggers logs.Loggers) *logStreamer {
	return &logStreamer{
		IsStdErr: isStdErr,
		Loggers:  loggers,
	}
}

func newOutStreamer(loggers logs.Loggers) *logStreamer {
	return newLogStreamer(false, loggers)
}

func newErrLogStreamer(loggers logs.Loggers) *logStreamer {
	return newLogStreamer(true, loggers)
}
