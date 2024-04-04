/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"context"
	"io"
	"strings"

	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/platform"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

var lineSep = platform.UnixLineSeparator()

// INTERNAL
// Way of redirecting process output to a logger.
type logStreamer struct {
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

func newLogStreamer(ctx context.Context, isStdErr bool, loggers logs.Loggers) io.Writer {
	return safeio.ContextualWriter(ctx, &logStreamer{
		IsStdErr: isStdErr,
		Loggers:  loggers,
	})
}

func newOutStreamer(ctx context.Context, loggers logs.Loggers) io.Writer {
	return newLogStreamer(ctx, false, loggers)
}

func newErrLogStreamer(ctx context.Context, loggers logs.Loggers) io.Writer {
	return newLogStreamer(ctx, true, loggers)
}
