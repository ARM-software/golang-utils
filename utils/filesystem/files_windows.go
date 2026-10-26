/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build windows

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package filesystem describes the filesystem on windows
package filesystem

import (
	"os"
	"syscall"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func isPrivilegeError(err error) bool {
	return commonerrors.Any(err, syscall.EPERM, syscall.ERROR_PRIVILEGE_NOT_HELD)
}

func isNotSupportedError(err error) bool {
	return commonerrors.Any(err, syscall.ENOTSUP, syscall.EOPNOTSUPP, syscall.EWINDOWS)
}

func determineFileOwners(_ os.FileInfo) (uid, gid int, err error) {
	uid = syscall.Getuid()
	gid = syscall.Getgid()

	return
}
