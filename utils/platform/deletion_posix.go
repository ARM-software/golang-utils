//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

func removeFileAs(ctx context.Context, as *command.CommandAsDifferentUser, path string) error {
	return executeCommandAs(ctx, as, "rm", "-f")
}

func removeDirAs(ctx context.Context, as *command.CommandAsDifferentUser, path string) error {
	return executeCommandAs(ctx, as, "rm", "-r", "-f")
}
