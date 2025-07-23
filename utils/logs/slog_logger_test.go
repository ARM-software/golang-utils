/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestSlogLogger(t *testing.T) {
	defer goleak.VerifyNone(t)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	loggers, err := NewSlogLogger(logger, "Test")
	require.NoError(t, err)
	testLog(t, loggers)
}
