/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package sharedcache

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestDefaultSharedCacheConfiguration(t *testing.T) {
	cfg := DefaultSharedCacheConfiguration()
	require.Error(t, cfg.Validate())
	cfg.RemoteStoragePath = faker.URL()
	require.NoError(t, cfg.Validate())
}
