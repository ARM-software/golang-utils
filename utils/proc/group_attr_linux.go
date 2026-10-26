/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build linux

package proc

import (
	"os/exec"
	"syscall"
)

// See https://github.com/tgulacsi/go/blob/master/proc/proc_linux.go
func setGroupAttrToCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}
}
