/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStdLogger(t *testing.T) {
	loggers, err := NewStdLogger("Test")
	require.NoError(t, err)
	testLog(t, loggers)
}

func TestAsynchronousStdLogger(t *testing.T) {
	loggers, err := NewAsynchronousStdLogger("Test", 1024, 2*time.Millisecond, "test source")
	require.NoError(t, err)
	testLog(t, loggers)
}
