/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
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

func TestGetLogrFromEmptyContext(t *testing.T) {
	ctx := context.Background()
	logger, err := GetLogrLoggerFromContext(ctx)

	assert.Equal(t, logr.Logger{}, logger)
	errortest.AssertError(t, err, commonerrors.ErrNoLogger)
}

func TestGetLogrFromContext(t *testing.T) {
	ctx := context.Background()
	logger := logstest.NewTestLogger(t)
	ctx = logr.NewContext(ctx, logger)

	newLogger, err := GetLogrLoggerFromContext(ctx)
	assert.NoError(t, err)

	assert.Equal(t, logger, newLogger)
}
