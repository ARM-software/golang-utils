/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build !windows

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireUserMutationPrivileges(t *testing.T) {
	t.Helper()
	root, err := IsCurrentWithRootPrivileges()
	require.NoError(t, err)
	if !root {
		t.Skip("Skipping user and group mutation tests without root privileges")
	}
}

func testUserPassword() string {
	return ""
}

func requireUserMutationNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}
