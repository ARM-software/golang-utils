/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package proc

import (
	"fmt"
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
		return fmt.Errorf("%w: %v", commonerrors.ErrTimeout, err.Error())
	case commonerrors.Any(err, exec.ErrDot, exec.ErrNotFound):
		return fmt.Errorf("%w: %v", commonerrors.ErrNotFound, err.Error())
	case commonerrors.Any(process.ErrorNotPermitted):
		return fmt.Errorf("%w: %v", commonerrors.ErrForbidden, err.Error())
	case commonerrors.Any(process.ErrorProcessNotRunning):
		return fmt.Errorf("%w: %v", commonerrors.ErrNotFound, err.Error())
	case commonerrors.CorrespondTo(err, errAccessDenied):
		return fmt.Errorf("%w: %v", commonerrors.ErrNotFound, err.Error())
	case commonerrors.CorrespondTo(err, errNotImplemented):
		return fmt.Errorf("%w: %v", commonerrors.ErrNotImplemented, err.Error())
	default:
		return err
	}
}
