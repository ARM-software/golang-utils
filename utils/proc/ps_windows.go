//go:build windows
// +build windows

//

/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/ARM-software/golang-utils/utils/collection"
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
	if commonerrors.Any(err, nil, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
		return
	} else {
		err = fmt.Errorf("%w: could not kill process group (#%v): %v", commonerrors.ErrUnexpected, pid, err.Error())
	}
	return
}

// WaitForCompletion will wait for a given process to complete.
// This allows check to work if the underlying process was stopped without needing the os.Process that started it.
func WaitForCompletion(ctx context.Context, pid int) (err error) {
	parent, err := FindProcess(ctx, pid)
	children, err := parent.Children(ctx)

	// Windows doesn't have group PIDs
	var pids = make([]int, len(children)+1)
	pids[0] = parent.Pid()
	for i := range children {
		pids[i+1] = children[i].Pid()
	}

	return parallelisation.WaitUntil(ctx, func(ctx2 context.Context) (bool, error) {
		return collection.AnyFunc(pids, func(pid int) bool {
			p, _ := FindProcess(ctx2, pid)
			return p.IsRunning()
		}), nil
	}, 1000*time.Millisecond)
}
