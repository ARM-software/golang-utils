//go:build darwin

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"os/exec"
	"syscall"
)

// See https://github.com/tgulacsi/go/blob/master/proc/proc_darwin.go
func setGroupAttrToCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
