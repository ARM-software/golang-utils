/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package proc

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/shirou/gopsutil/v4/process"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

const (
	errKilledProcess     = "signal: killed"
	errTerminatedProcess = "signal: terminated"
	errAccessDenied      = "Access is denied"
	errNotImplemented    = "not implemented"
)

func ConvertProcessError(err error) error {
	err = commonerrors.ConvertContextError(err)
	switch {
	case err == nil:
		return err
	case commonerrors.CorrespondTo(err, errKilledProcess, errTerminatedProcess):
		return os.ErrProcessDone
	case commonerrors.Any(err, syscall.ESRCH):
		// ESRCH is "no such process", meaning the process has already exited.
		return nil
	case commonerrors.Any(err, exec.ErrWaitDelay):
		return commonerrors.WrapError(commonerrors.ErrTimeout, err, "")
	case commonerrors.Any(err, exec.ErrDot, exec.ErrNotFound):
		return commonerrors.WrapError(commonerrors.ErrNotFound, err, "")
	case commonerrors.Any(process.ErrorNotPermitted):
		return commonerrors.WrapError(commonerrors.ErrForbidden, err, "")
	case commonerrors.Any(process.ErrorProcessNotRunning):
		return commonerrors.WrapError(commonerrors.ErrNotFound, err, "")
	case commonerrors.CorrespondTo(err, errAccessDenied):
		return commonerrors.WrapError(commonerrors.ErrNotFound, err, "")
	case commonerrors.CorrespondTo(err, errNotImplemented):
		return commonerrors.WrapError(commonerrors.ErrNotImplemented, err, "")
	default:
		return err
	}
}
