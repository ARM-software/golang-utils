/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build windows

package platform

import "testing"

func requireUserMutationPrivileges(t *testing.T) {
	t.Helper()
	t.Log("Note: it is necessary to run this test with elevated privileges https://github.com/iamacarpet/go-win64api/issues/26")
	admin, _ := IsCurrentUserAnAdmin()
	if !admin {
		t.Skip("Skipping as running on windows with a non-admin user")
	}
}
