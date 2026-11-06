/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build !windows

package platform

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertCurrentUserAdminState(t *testing.T, currentUser *user.User) {
	t.Helper()
	admin, err := IsAdmin(currentUser.Username)
	require.NoError(t, err)
	adminFromUser, err := IsUserAdmin(currentUser)
	require.NoError(t, err)
	assert.Equal(t, admin, adminFromUser)
	currentAdmin, err := IsCurrentUserAnAdmin()
	require.NoError(t, err)
	assert.Equal(t, currentAdmin, adminFromUser)
}
