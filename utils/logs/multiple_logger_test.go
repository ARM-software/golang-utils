/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
)

func TestMultipleLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggers, err := NewMultipleLoggers("Test")
	require.NoError(t, err)
	testLog(t, loggers)
}

func TestCombinedLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	_, err := NewCombinedLoggers()
	errortest.RequireError(t, err, commonerrors.ErrNoLogger)
	testLogger, err := NewLogrLogger(logstest.NewTestLogger(t), "Test")
	require.NoError(t, err)
	nl, err := NewNoopLogger("Test2")
	require.NoError(t, err)
	loggers, err := NewCombinedLoggers(testLogger, nl)
	require.NoError(t, err)
	testLog(t, loggers)
}

func TestMultipleLoggers(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("Manually add loggers", func(t *testing.T) {
		// With default logger
		loggers, err := NewMultipleLoggers("Test Multiple")
		require.NoError(t, err)
		testLog(t, loggers)

		// Adding a file logger to the mix.
		file, err := filesystem.TempFileInTempDir("test-multiplelog-filelog-*.log")
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		defer func() { _ = filesystem.Rm(file.Name()) }()

		empty, err := filesystem.IsEmpty(file.Name())
		require.NoError(t, err)
		assert.True(t, empty)

		fl, err := NewFileLogger(file.Name(), "Test")
		require.NoError(t, err)

		require.NoError(t, loggers.Append(fl))

		nl, err := NewNoopLogger("Test2")
		require.NoError(t, err)

		// Adding various loggers
		require.NoError(t, loggers.Append(fl, nl))

		testLog(t, loggers)

		empty, err = filesystem.IsEmpty(file.Name())
		require.NoError(t, err)
		assert.False(t, empty)

		err = filesystem.Rm(file.Name())
		require.NoError(t, err)

		// Concurrency test multiple loggers

		stdLogger, err := NewStdLogger("Test std logger")
		require.NoError(t, err)

		stringLogger, err := NewPlainStringLogger()
		if err != nil {
			return
		}
		mLoggers, err := NewCombinedLoggers(stdLogger, stringLogger)
		if err != nil {
			return
		}

		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				mLoggers.Log(fmt.Sprintf("Test output %v", i))
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				mLoggers.LogError(fmt.Sprintf("Test output %v", i))
			}
		}()

		wg.Wait()
		err = loggers.Close()
		require.NoError(t, err)
	})

	t.Run("Add loggers at start", func(t *testing.T) {
		file, err := filesystem.TempFileInTempDir("test-multiplelog-filelog-*.log")
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		defer func() { _ = filesystem.Rm(file.Name()) }()

		empty, err := filesystem.IsEmpty(file.Name())
		require.NoError(t, err)
		assert.True(t, empty)

		fl, err := NewFileLogger(file.Name(), "Test")
		require.NoError(t, err)

		nl, err := NewNoopLogger("Test2")
		require.NoError(t, err)

		loggers, err := NewMultipleLoggers("Test Multiple", fl, nl)
		require.NoError(t, err)
		testLog(t, loggers)
	})
}
