/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestLogMessage(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("with writer closing", func(t *testing.T) {
		loggers, err := NewJSONLogger(NewStdWriterWithSource(), "Test", "TestLogMessage")
		require.NoError(t, err)
		testLog(t, loggers)
	})
	t.Run("without writer closing", func(t *testing.T) {
		writer := NewStdWriterWithSource()
		defer func() { _ = writer.Close() }()
		loggers, err := NewJSONLoggerWithWriter(writer, "Test", "TestLogMessage")
		require.NoError(t, err)
		testLog(t, loggers)
		require.NoError(t, writer.Close())
	})
}

func TestLogMessageToSlowLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	stdloggers, err := NewStdLogger("ERR:")
	require.NoError(t, err)

	defer goleak.VerifyNone(t)
	t.Run("with writer closing", func(t *testing.T) {
		loggers, err := NewJSONLoggerForSlowWriter(NewTestSlowWriter(t), 1024, 2*time.Millisecond, "Test", "TestLogMessageToSlowLogger", stdloggers)
		require.NoError(t, err)
		testLog(t, loggers)
		time.Sleep(100 * time.Millisecond)
	})
	t.Run("without writer closing", func(t *testing.T) {
		writer := NewTestSlowWriter(t)
		defer func() { _ = writer.Close() }()
		loggers, err := NewJSONLoggerForSlowWriters(writer, 1024, 2*time.Millisecond, "Test", "TestLogMessageToSlowLogger", stdloggers)
		require.NoError(t, err)
		testLog(t, loggers)
		time.Sleep(100 * time.Millisecond)
		require.NoError(t, writer.Close())
	})
}
