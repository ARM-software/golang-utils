/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestStringLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggers, err := NewStringLogger("Test")
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError("Test err")
	loggers.Log("Test1")
	contents := loggers.GetLogContent()
	require.NotEmpty(t, contents)
	require.True(t, strings.Contains(contents, "Test err"))
	require.True(t, strings.Contains(contents, "Test1"))
}

func TestPlainStringLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggers, err := NewPlainStringLogger()
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError("Test err")
	loggers.Log("Test1")
	contents := loggers.GetLogContent()
	require.NotEmpty(t, contents)
	require.True(t, strings.Contains(contents, "Test err"))
	require.True(t, strings.Contains(contents, "Test1"))
}
