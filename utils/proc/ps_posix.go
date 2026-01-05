//go:build !windows

/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"context"
	"syscall"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

func getpgid(pid int) (gpid int, err error) {
	gpid, err = syscall.Getpgid(pid)
	switch {
	case commonerrors.CorrespondTo(err, "no such process"):
		err = commonerrors.Newf(commonerrors.ErrNotFound, "process '%v' does not exist", pid)
		return
	case err != nil:
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not get pgid of '%v'", pid)
		return
	default:
		return
	}
}

func killGroup(ctx context.Context, pid int32) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	// see https://varunksaini.com/posts/kiling-processes-in-go/
	pgid, err := getpgid(int(pid))
	if err != nil {
		return
	}
	// kill a whole process group by sending a signal to -xyz where xyz is the pgid
	// http://unix.stackexchange.com/questions/14815/process-descendants
	if pgid != int(pid) {
		err = commonerrors.Newf(commonerrors.ErrUnexpected, "process #%v is not group leader", pid)
		return
	}
	err = ConvertProcessError(syscall.Kill(-pgid, syscall.SIGKILL))
	return
}

func getGroupProcesses(ctx context.Context, pid int) (pids []int, err error) {
	pgid, err := getpgid(pid)
	switch {
	case commonerrors.Any(err, commonerrors.ErrNotFound):
		return
	case err != nil:
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not get group PID for '%v'", pid)
		return
	default:
		pids = append(pids, pgid)
		return
	}
}
