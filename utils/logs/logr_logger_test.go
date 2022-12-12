/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
)

func TestLogrLogger(t *testing.T) {
	loggers, err := NewLogrLogger(logstest.NewTestLogger(t), "Test")
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError(commonerrors.ErrUnexpected, ": no idea what happened")
	loggers.LogError(nil, ": no idea what happened")
	loggers.LogError("no idea what happened")
	loggers.LogError("no idea what happened", nil)
}

func TestLogrLoggerConversion(t *testing.T) {
	loggers, err := NewLogrLogger(logstest.NewTestLogger(t), "Test")
	require.NoError(t, err)
	converted := NewLogrLoggerFromLoggers(loggers)
	converted.WithName(faker.Name()).WithValues(faker.Word(), faker.Name()).Error(commonerrors.ErrUnexpected, faker.Sentence())
}
