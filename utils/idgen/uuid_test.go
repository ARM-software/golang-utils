/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package idgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUuidUniqueness(t *testing.T) {
	uuid1, err := GenerateUUID4()
	require.NoError(t, err)

	uuid2, err := GenerateUUID4()
	require.NoError(t, err)

	assert.NotEqual(t, uuid1, uuid2)
}

func TestUuidLength(t *testing.T) {
	uuid, err := GenerateUUID4()
	require.NoError(t, err)

	assert.Equal(t, 36, len(uuid))
}

func TestInvalidUUID(t *testing.T) {
	id := "1"
	require.False(t, IsValidUUID(id))
}

func TestValidUUID(t *testing.T) {
	id := "5a4b6bb3-0bd3-4c4e-ba4c-45658cca4289"
	require.True(t, IsValidUUID(id))
}
