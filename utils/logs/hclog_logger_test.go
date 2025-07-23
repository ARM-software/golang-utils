/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/logs/logstest"
)

func TestHclogLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	logger := hclog.New(nil)
	loggers, err := NewHclogLogger(logger, "Test")
	require.NoError(t, err)
	testLog(t, loggers)
}

func TestHclogWrapper(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggers, err := NewLogrLogger(logstest.NewTestLogger(t), "test")
	require.NoError(t, err)
	hcLogger, err := NewHclogWrapper(loggers)
	require.NoError(t, err)
	loggerTest, err := NewHclogLogger(hcLogger, "Test")
	require.NoError(t, err)
	testLog(t, loggerTest)
}
