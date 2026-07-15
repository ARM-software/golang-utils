/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package retry

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.False(t, configTest.RetryAfterDisabled)
}

func TestFixedBackoffRetryConfiguration(t *testing.T) {
	configTest := DefaultFixedBackoffRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
	assert.True(t, configTest.Enabled)
	assert.True(t, configTest.BackOffEnabled)
	assert.True(t, configTest.LinearBackOffEnabled)
	assert.Equal(t, configTest.RetryWaitMin, configTest.RetryWaitMax)
	assert.Zero(t, configTest.RetryMaxJitter)
}

func TestRobustFixedBackoffRetryConfiguration(t *testing.T) {
	configTest := DefaultRobustFixedBackoffRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
	assert.False(t, configTest.RetryAfterDisabled)
	assert.True(t, configTest.Enabled)
}

func TestExponentialBackoffRetryConfiguration(t *testing.T) {
	configTest := DefaultExponentialBackoffRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestLinearBackoffRetryConfiguration(t *testing.T) {
	configTest := DefaultLinearBackoffRetryPolicyConfiguration()
	require.NoError(t, configTest.Validate())
}
