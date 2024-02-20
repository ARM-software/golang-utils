/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package retry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoRetryConfiguration(t *testing.T) {
	configTest := DefaultNoRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestBasicRetryConfiguration(t *testing.T) {
	configTest := DefaultBasicRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestBasicRetryWithRetryAfterConfiguration(t *testing.T) {
	configTest := DefaultRobustRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestExponentialBackoffRetryConfiguration(t *testing.T) {
	configTest := DefaultExponentialBackoffRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestLinearBackoffRetryConfiguration(t *testing.T) {
	configTest := DefaultLinearBackoffRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
}
