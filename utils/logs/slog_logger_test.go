/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestSlogLogger(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	loggers, err := NewSlogLogger(logger, "Test")
	require.NoError(t, err)
	testLog(t, loggers)
}
