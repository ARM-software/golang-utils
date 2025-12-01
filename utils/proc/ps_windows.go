//go:build windows

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

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

func killGroup(ctx context.Context, pid int32) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	cmd := exec.CommandContext(ctx, "taskkill", "/f", "/t", "/pid", strconv.Itoa(int(pid)))
	// setting the following to avoid having hanging subprocesses as described in https://github.com/golang/go/issues/24050
	cmd.WaitDelay = 50 * time.Millisecond
	err = ConvertProcessError(cmd.Run())
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not kill process group (#%v)", pid)
	}
	return
}

func getGroupProcesses(ctx context.Context, pid int) (pids []int, err error) {
	parent, err := FindProcess(ctx, pid)
	if err != nil {
		return
	}
	children, err := parent.Children(ctx)
	if err != nil {
		return
	}
	// Windows doesn't have group PIDs
	pids = make([]int, len(children)+1)
	pids[0] = parent.Pid()
	for i := range children {
		pids[i+1] = children[i].Pid()
	}
	return
}
