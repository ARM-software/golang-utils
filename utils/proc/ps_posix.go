//go:build !windows
// +build !windows

/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

func killGroup(ctx context.Context, pid int32) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	// see https://varunksaini.com/posts/kiling-processes-in-go/
	pgid, err := syscall.Getpgid(int(pid))
	if err != nil {
		return
	}
	// kill a whole process group by sending a signal to -xyz where xyz is the pgid
	// http://unix.stackexchange.com/questions/14815/process-descendants
	if pgid != int(pid) {
		err = fmt.Errorf("%w: process #%v is not group leader", commonerrors.ErrUnexpected, pid)
		return
	}
	err = ConvertProcessError(syscall.Kill(-pgid, syscall.SIGKILL))
	return
}

// WaitForCompletion will wait for a given process to complete.
// This allows check to work if the underlying process was stopped without needing the os.Process that started it.
func WaitForCompletion(ctx context.Context, pid int) (err error) {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not get group PID for '%v'", pid)
		return
	}

	return parallelisation.WaitUntil(ctx, func(ctx2 context.Context) (bool, error) {
		p, _ := FindProcess(ctx, pgid)
		return p.IsRunning(), nil // FindProcess will always return an instantiated process and any non-runnning state should exit without error
	}, 1000*time.Millisecond)
}
