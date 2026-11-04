//go:build windows

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"fmt"
	"os/exec"
	"time"
)

func sleepCommand(duration time.Duration) *exec.Cmd {
	return exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", fmt.Sprintf("Start-Sleep -Milliseconds %d", duration.Milliseconds()))
}
