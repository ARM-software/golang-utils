/*
 * Copyright (C) 2020-2023 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuietLogger(t *testing.T) {
	logger, err := NewStdLogger("Test")
	require.NoError(t, err)
	loggers, err := NewQuietLogger(logger)
	require.NoError(t, err)
	testLog(t, loggers)
}
