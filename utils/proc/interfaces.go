/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package proc

import "context"

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IProcess

// IProcess is the generic interface that is implemented on every platform
// and provides common operations for processes.
// Inspired from https://github.com/mitchellh/go-ps/blob/master/process.go
type IProcess interface {
	// Pid is the process ID for this process.
	Pid() int

	// PPid is the parent process ID for this process.
	PPid() int

	// Executable name running this process. This is not a path to the
	// executable.
	Executable() string

	// Name returns name of the process.
	Name() string

	Environ(ctx context.Context) ([]string, error)

	// Cmdline returns the command line arguments of the process as a string with
	// each argument separated by 0x20 ascii character.
	Cmdline() string

	// Cwd returns current working directory of the process.
	Cwd() string

	// Parent returns parent Process of the process.
	Parent() IProcess

	// Children returns the children of the process if any.
	Children(ctx context.Context) ([]IProcess, error)

	// IsRunning returns whether the process is still running or not.
	IsRunning() bool

	// Terminate sends SIGTERM to the process.
	Terminate(context.Context) error

	// Interrupt sends SIGINT to the process.
	Interrupt(context.Context) error

	// KillWithChildren sends SIGKILL to the process but also ensures any children of the process are also killed.
	// see https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	// This method was introduced to avoid getting the following due to a poor cleanup:
	// - Orphan processes (https://en.wikipedia.org/wiki/Orphan_process)
	// - Zombies processes (https://en.wikipedia.org/wiki/Zombie_process)
	KillWithChildren(context.Context) error
}
