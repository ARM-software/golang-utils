/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package proc

import "os/exec"

// See https://github.com/tgulacsi/go/blob/master/proc/
func SetGroupAttrToCmd(c *exec.Cmd) {
	setGroupAttrToCmd(c)
}
