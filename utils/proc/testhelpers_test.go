/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package proc

import "os/exec"

// newCommandWithChildProcess returns a portable bash command that stays alive
// while also spawning a child process, so process-tree termination behaviour
// can be exercised in tests.
// See https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
// and https://forum.golangbridge.org/t/killing-child-process-on-timeout-in-go-code/995/16.
func newCommandWithChildProcess() *exec.Cmd {
	return exec.Command("bash", "-c", "sleep 50 & wait")
}
