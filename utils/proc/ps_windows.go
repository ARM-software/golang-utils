//go:build windows
// +build windows

/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"context"
	"os/exec"
	"strconv"
	"time"

	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

func killGroup(ctx context.Context, pid int32) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	cmd := exec.CommandContext(ctx, "taskkill", "/f", "/t", "/pid", strconv.Itoa(int(pid)))
	// setting the following to avoid having hanging subprocesses as described in https://github.com/golang/go/issues/24050
	cmd.WaitDelay = 500 * time.Millisecond
	return ConvertProcessError(cmd.Run())

}
