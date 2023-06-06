/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestHclogLogger(t *testing.T) {
	logger := hclog.New(nil)
	loggers, err := NewHclogLogger(logger, "Test")
	require.NoError(t, err)
	testLog(t, loggers)
}
