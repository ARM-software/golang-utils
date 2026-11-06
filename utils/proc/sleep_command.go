/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package proc

import (
	"os/exec"
	"time"
)

// SleepCommand returns a command that sleeps for the requested duration using
// the current platform's native implementation.
func SleepCommand(duration time.Duration) *exec.Cmd {
	return sleepCommand(duration)
}
