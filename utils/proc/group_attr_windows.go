//go:build windows

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"os/exec"
	"syscall"
)

// See https://github.com/tgulacsi/go/blob/master/proc/proc_windows.go
func setGroupAttrToCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		// Windows Process Creation Flags:  https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
		CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT | syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
