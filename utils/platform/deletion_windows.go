//go:build windows

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
	err := executeCommandAs(ctx, as, "takeown", "/f", path)
	if err != nil {
		return err
	}
	return executeCommandAs(ctx, as, "del", "/q", "/f", path)
}

func removeDirAs(ctx context.Context, as *command.CommandAsDifferentUser, path string) error {
	err := executeCommandAs(ctx, as, "takeown", "/r", "/d", "Y", "/f", path)
	if err != nil {
		return err
	}
	return executeCommandAs(ctx, as, "rmdir", "/s", "/q", path)
}
