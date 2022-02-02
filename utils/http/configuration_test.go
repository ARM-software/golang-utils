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
	require.Nil(t, configTest.Validate())
}

func TestFastHttpClientConfiguration(t *testing.T) {
	configTest := FastHTTPClientConfiguration()
	require.Nil(t, configTest.Validate())
}
