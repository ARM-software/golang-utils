/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build !windows

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"os/exec"
	"strconv"
	"time"
)

func sleepCommand(duration time.Duration) *exec.Cmd {
	return exec.Command("sleep", strconv.FormatFloat(duration.Seconds(), 'f', -1, 64)) //nolint:gosec // duration is typed input converted to a numeric argument, not user-controlled shell input
}
