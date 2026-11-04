/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build windows

package platform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireUserMutationPrivileges(t *testing.T) {
	t.Helper()
	t.Log("Note: it is necessary to run this test with elevated privileges https://github.com/iamacarpet/go-win64api/issues/26")
	allowed, err := CanCurrentUserMutateUsers()
	require.NoError(t, err)
	if !allowed {
		t.Skip("Skipping as running on windows with a non-admin user")
	}
}

func testUserPassword() string {
	return "T3stPassw0rd!"
}

func requireUserMutationNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}
