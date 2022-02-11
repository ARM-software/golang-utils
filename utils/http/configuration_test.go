/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpClientConfiguration(t *testing.T) {
	configTest := DefaultHTTPClientConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestFastHttpClientConfiguration(t *testing.T) {
	configTest := FastHTTPClientConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestHttpClientConfigurationWithRetry(t *testing.T) {
	configTest := DefaultRobustHTTPClientConfiguration()
	require.NoError(t, configTest.Validate())
}

func TestHttpClientConfigurationWithRetryAndRetryAfter(t *testing.T) {
	configTest := DefaultRobustHTTPClientConfigurationWithRetryAfter()
	require.NoError(t, configTest.Validate())
}

func TestHttpClientConfigurationWithBackoff(t *testing.T) {
	configTest := DefaultRobustHTTPClientConfigurationWithExponentialBackOff()
	require.NoError(t, configTest.Validate())
}

func TestHttpClientConfigurationWithLinearBackoff(t *testing.T) {
	configTest := DefaultRobustHTTPClientConfigurationWithLinearBackOff()
	require.NoError(t, configTest.Validate())
}
