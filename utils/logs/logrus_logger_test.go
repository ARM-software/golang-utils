/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package logs

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLogrusLogger(t *testing.T) {
	loggers, err := NewLogrusLogger(logrus.StandardLogger(), "Test")
	require.Nil(t, err)
	testLog(t, loggers)
}
