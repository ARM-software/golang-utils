/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopLogger(t *testing.T) {
	loggers, err := NewNoopLogger("Test")
	require.NoError(t, err)
	testLog(t, loggers)
}
