/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build linux

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package subprocess

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs"
)

const interactiveTerminalHelperTest = "TestInteractiveTerminalProcess"

func TestExecuteInteractiveWithDirTerminal(t *testing.T) {
	const input = "terminal input\n"

	for _, mode := range []string{
		"read",
		"cancel",
	} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			controller, terminal := createInteractivePTY(t)
			drainInteractiveTerminal(t, controller)

			command, args := interactiveHelperCommand(t, interactiveTerminalHelperTest, "driver", dir, mode)
			driver := exec.Command(command, args...)
			driver.Stdin, driver.Stdout, driver.Stderr = terminal, terminal, terminal
			driver.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}

			done := startInteractiveTestProcess(t, driver)
			require.NoError(t, terminal.Close())
			driverFinished := false

			if mode == "read" {
				childPID := waitInteractiveEvent(t, dir, "read.ready")
				require.Greater(t, childPID, 1)
				require.NotEqual(t, os.Getpid(), childPID)
				require.NotEqual(t, driver.Process.Pid, childPID)

				t.Cleanup(func() {
					if driverFinished {
						return
					}

					// SIGTTIN can stop a regressed child outside the driver's group.
					_ = unix.Kill(childPID, unix.SIGKILL)

					select {
					case exitErr := <-done:
						if exitErr != nil {
							t.Logf("terminal driver exit during cleanup: %v", exitErr)
						}
					case <-time.After(time.Second): // Allow the driver to reap before its cleanup.
					}
				})

				driverGroup, err := unix.Getpgid(driver.Process.Pid)
				require.NoError(t, err)
				childGroup, err := unix.Getpgid(childPID)
				require.NoError(t, err)

				require.Equal(t, driver.Process.Pid, driverGroup, "driver must lead its isolated group")
				require.Equal(t, driverGroup, childGroup, "interactive child must inherit the driver's process group")
				require.Equal(t, interactiveTerminalForeground(t, controller), childGroup, "interactive child must be in the terminal foreground group")

				_, err = controller.WriteString(input)
				require.NoError(t, err)
			}

			waitInteractiveExit(t, done)
			driverFinished = true

			assert.Equal(t, driver.Process.Pid, waitInteractiveEvent(t, dir, "driver.done"))

			if mode == "read" {
				data, readErr := filesystem.ReadFile(filepath.Join(dir, "read.input"))
				require.NoError(t, readErr)
				assert.Equal(t, []byte(input), data, "the child must read the exact PTY input")
			}
		})
	}
}

func createInteractivePTY(t *testing.T) (controller, terminal *os.File) {
	t.Helper()

	controllerFD, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if errors.Is(err, unix.ENOENT) || errors.Is(err, unix.ENODEV) || errors.Is(err, unix.ENXIO) || errors.Is(err, unix.EACCES) || errors.Is(err, unix.EPERM) {
		t.Skipf("Linux PTY unavailable: open /dev/ptmx: %v", err)
	}

	require.NoError(t, err)
	controller = os.NewFile(uintptr(controllerFD), "/dev/ptmx")
	t.Cleanup(func() { _ = controller.Close() })

	// Non-blocking creation keeps the raw *os.File pollable, including deadlines.
	require.NoError(t, controller.SetDeadline(time.Now().Add(interactiveHelperTimeout)))
	require.NoError(t, unix.IoctlSetPointerInt(controllerFD, unix.TIOCSPTLCK, 0))
	terminalNumber, err := unix.IoctlGetInt(controllerFD, unix.TIOCGPTN)
	require.NoError(t, err)

	terminalPath := fmt.Sprintf("/dev/pts/%d", terminalNumber)
	terminalFD, err := unix.Open(terminalPath, unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	require.NoError(t, err)
	terminal = os.NewFile(uintptr(terminalFD), terminalPath)
	t.Cleanup(func() { _ = terminal.Close() })

	return
}

func interactiveTerminalForeground(t *testing.T, terminal *os.File) (group int) {
	t.Helper()

	// File.Fd would disable non-blocking mode and break the controller's deadlines.
	raw, err := terminal.SyscallConn()
	require.NoError(t, err)

	var ioctlErr error
	err = raw.Control(func(fd uintptr) { group, ioctlErr = unix.IoctlGetInt(int(fd), unix.TIOCGPGRP) })
	require.NoError(t, err)
	require.NoError(t, ioctlErr, "%s must be a controlling terminal", terminal.Name())

	return
}

func drainInteractiveTerminal(t *testing.T, controller *os.File) {
	t.Helper()

	done := make(chan string, 1)
	go func() {
		var output strings.Builder
		_, _ = io.Copy(&output, controller) // Linux PTYs normally end with EIO, not EOF.
		done <- output.String()
	}()

	// Registered before process cleanup so output is drained until the driver exits.
	t.Cleanup(func() {
		require.NoError(t, controller.SetReadDeadline(time.Now().Add(time.Second)))

		select {
		case output := <-done:
			if t.Failed() {
				t.Logf("terminal driver output:\n%s", output)
			}
		case <-time.After(interactiveTestTimeout):
			assert.Fail(t, "terminal output reader did not stop")
		}
	})
}

func TestInteractiveTerminalProcess(t *testing.T) {
	args := interactiveHelperArguments(t)
	if args == nil {
		return
	}

	for _, stream := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
		interactiveTerminalForeground(t, stream)
	}

	if args[0] != "driver" {
		runInteractiveHelper(t, args)
		return
	}

	require.Len(t, args, 3)
	dir := args[1]

	switch args[2] {
	case "read":
		logger, err := logs.NewNoopLogger("interactive-terminal")
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(t.Context(), interactiveTestTimeout)
		defer cancel()

		command, arguments := interactiveHelperCommand(t, interactiveTerminalHelperTest, "read", dir)
		require.NoError(t, ExecuteInteractiveWithDir(ctx, logger, "", "", "", dir, command, arguments...))
	case "cancel":
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		defer signal.Stop(interrupts)

		checkInteractiveCancellation(t, interactiveTerminalHelperTest, dir, interrupts)
	default:
		require.FailNowf(t, "unknown terminal driver scenario", "%q", args[2])
	}

	markInteractiveEvent(t, dir, "driver.done")
}
