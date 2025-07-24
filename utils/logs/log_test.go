/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestLog(t *testing.T) {
	defer goleak.VerifyNone(t)
	var loggers Loggers = &GenericLoggers{}
	err := loggers.Check()
	assert.Error(t, err)
	err = loggers.Close()
	assert.NoError(t, err)
}

func testLog(t *testing.T, loggers Loggers) {
	t.Helper()
	err := loggers.Check()
	require.NoError(t, err)
	defer func() { _ = loggers.Close() }()

	err = loggers.SetLogSource("source1")
	require.NoError(t, err)
	err = loggers.SetLoggerSource("LoggerSource1")
	require.NoError(t, err)

	loggers.Log("Test output1")
	loggers.Log("Test output2")
	loggers.Log("\"/usr/bin/armlink\" --via=\"/workspace/Objects/aws_mqtt_demo.axf._ld\"\n")
	loggers.Log("\n")
	loggers.LogError("\n")
	err = loggers.SetLogSource("source2")
	require.NoError(t, err)

	loggers.Log("Test output3")
	loggers.LogError("Test err1")
	err = loggers.SetLogSource("source3")
	require.NoError(t, err)

	err = loggers.SetLoggerSource("LoggerSource2")
	require.NoError(t, err)

	loggers.LogError("Test err2")
	err = loggers.SetLogSource("source4")
	require.NoError(t, err)

	loggers.LogError("Test err3")
	loggers.LogError(commonerrors.ErrCancelled)
	loggers.LogError(nil)
	loggers.LogError(commonerrors.ErrUnexpected, "some error")
	loggers.LogError("some error", commonerrors.ErrUnexpected)
	loggers.LogError(nil, "no error")
	err = loggers.Close()
	require.NoError(t, err)
}
