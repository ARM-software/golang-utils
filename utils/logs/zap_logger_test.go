/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"
)

func TestZapLoggerDev(t *testing.T) {
	defer goleak.VerifyNone(t)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	loggers, err := NewZapLogger(logger, "Test")
	require.NoError(t, err)
	testLog(t, loggers)
}

func TestZapLoggerProd(t *testing.T) {
	defer goleak.VerifyNone(t)
	logger, err := zap.NewProduction()
	require.NoError(t, err)
	loggers, err := NewZapLogger(logger, "Test")
	require.NoError(t, err)
	testLog(t, loggers)
}
