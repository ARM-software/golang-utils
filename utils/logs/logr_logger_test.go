/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
)

func TestLogrLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggers, err := NewLogrLogger(logstest.NewTestLogger(t), "Test")
	require.NoError(t, err)
	testLog(t, loggers)
	loggers.LogError(commonerrors.ErrUnexpected, ": no idea what happened")
	loggers.LogError(nil, ": no idea what happened")
	loggers.LogError("no idea what happened")
	loggers.LogError("no idea what happened", nil)
}

func TestLogrLoggerConversion(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggers, err := NewLogrLogger(logstest.NewTestLogger(t), "Test")
	require.NoError(t, err)
	converted := NewLogrLoggerFromLoggers(loggers)
	converted.WithName(faker.Name()).WithValues(faker.Word(), faker.Name()).Error(commonerrors.ErrUnexpected, faker.Sentence())
	converted.Info(faker.Sentence(), faker.Word(), faker.Name())
}

func TestLogrFromLoggersWithMultipleName(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggerSource := "src-" + faker.Name()
	strLogger, err := NewStringLogger(loggerSource)
	require.NoError(t, err)
	converted := NewLogrLoggerFromLoggers(strLogger)
	converted.WithName(loggerSource).WithName(faker.Name()).WithName(faker.Name()).WithName(faker.Name()).WithName(faker.Name()).WithName(faker.Name()).WithName(loggerSource).WithName(faker.Name()).Error(commonerrors.ErrUnexpected, faker.Sentence())
	assert.Contains(t, strLogger.GetLogContent(), loggerSource)
	assert.Equal(t, 2, strings.Count(strLogger.GetLogContent(), loggerSource))
	converted.WithName(loggerSource).Info(faker.Sentence(), faker.Word(), faker.Name())
	assert.Equal(t, 4, strings.Count(strLogger.GetLogContent(), loggerSource))
	fmt.Println(strLogger.GetLogContent())
}

func TestLoggersFromLoggerWithMultipleSource(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggerSource := "src-" + faker.Name()
	strLogger, err := NewStringLogger(loggerSource)
	require.NoError(t, err)
	logrl := NewLogrLoggerFromLoggers(strLogger)
	logger, err := NewLogrLogger(logrl, loggerSource)
	require.NoError(t, err)
	require.NoError(t, logger.SetLogSource("logsrc-"+faker.Name()))
	require.NoError(t, logger.SetLoggerSource(loggerSource))
	require.NoError(t, logger.SetLoggerSource(loggerSource))
	require.NoError(t, logger.SetLoggerSource(faker.Name()))
	require.NoError(t, logger.SetLoggerSource(loggerSource))
	logger.Log(faker.Sentence())
	assert.Contains(t, strLogger.GetLogContent(), loggerSource)
	assert.Equal(t, 2, strings.Count(strLogger.GetLogContent(), loggerSource))
	require.NoError(t, logger.SetLoggerSource(loggerSource))
	require.NoError(t, logger.SetLoggerSource(faker.Name()))
	logger.Log(faker.Sentence())
	assert.Contains(t, strLogger.GetLogContent(), loggerSource)
	assert.Equal(t, 4, strings.Count(strLogger.GetLogContent(), loggerSource))
}

func TestLogrLoggerConversionPlain(t *testing.T) {
	defer goleak.VerifyNone(t)
	loggers, err := NewPipeLogger()
	require.NoError(t, err)
	converted := NewPlainLogrLoggerFromLoggers(loggers)
	converted.WithName(faker.Name()).WithValues(faker.Word(), faker.Name()).Error(commonerrors.ErrUnexpected, faker.Sentence())
	converted.Info(faker.Sentence(), faker.Word(), faker.Name())
}

func TestGetLogrFromEmptyContext(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	logger, err := GetLogrLoggerFromContext(ctx)

	assert.Equal(t, logr.Logger{}, logger)
	errortest.AssertError(t, err, commonerrors.ErrNoLogger)
}

func TestGetLogrFromContext(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	logger := logstest.NewTestLogger(t)
	ctx = logr.NewContext(ctx, logger)

	newLogger, err := GetLogrLoggerFromContext(ctx)
	assert.NoError(t, err)

	assert.Equal(t, logger, newLogger)
}
