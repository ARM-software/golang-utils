/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

type SlowWriter struct {
	StdWriter
}

func (w *SlowWriter) Write(p []byte) (n int, err error) {
	time.Sleep(10 * time.Millisecond)
	return os.Stdout.Write(p)
}

func NewTestSlowWriter(t *testing.T) *SlowWriter {
	t.Helper()
	return &SlowWriter{}
}

// Creates a logger to standard output/error
func NewTestMultipleWriterLogger(t *testing.T, prefix string) (loggers Loggers) {
	t.Helper()
	writer, err := NewMultipleWritersWithSource(NewTestSlowWriter(t), NewTestSlowWriter(t))
	require.NoError(t, err)
	loggers = &GenericLoggers{
		Output: log.New(writer, "["+prefix+"] Output: ", log.LstdFlags),
		Error:  log.New(writer, "["+prefix+"] Error: ", log.LstdFlags),
	}
	return
}

func TestMultipleWriters(t *testing.T) {
	defer goleak.VerifyNone(t)
	testLog(t, NewTestMultipleWriterLogger(t, "Test"))
	time.Sleep(100 * time.Millisecond)
}
