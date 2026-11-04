//go:build windows

/*
 * Copyright (C) 2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package subprocess

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"

	"github.com/ARM-software/golang-utils/utils/logs"
)

func TestExecuteInteractiveWithDirWindowsConsole(t *testing.T) {
	for _, mode := range []string{
		"ctrl-c",
		"cancel",
	} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			binary, args := interactiveHelperCommand(t, "TestInteractiveWindowsHelperProcess", "driver", dir, mode)

			ctx, cancel := context.WithTimeout(t.Context(), interactiveHelperTimeout)
			defer cancel()

			// Console-wide Ctrl+C must never reach the main test runner.
			driver := exec.CommandContext(ctx, binary, args...)
			driver.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_CONSOLE, HideWindow: true}

			var output bytes.Buffer
			driver.Stdout, driver.Stderr = &output, &output
			t.Cleanup(func() {
				if t.Failed() {
					t.Logf("isolated console driver output:\n%s", output.String())
				}
			})

			done := startInteractiveTestProcess(t, driver)
			waitInteractiveExit(t, done)

			assert.Equal(t, driver.Process.Pid, waitInteractiveEvent(t, dir, "driver.done"))
		})
	}
}

func TestInteractiveWindowsHelperProcess(t *testing.T) {
	args := interactiveHelperArguments(t)
	if args == nil {
		return
	}

	if args[0] == "driver" {
		interactiveWindowsAttachConsole(t)
	}

	for _, stream := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
		var consoleMode uint32
		require.NoError(t, windows.GetConsoleMode(windows.Handle(stream.Fd()), &consoleMode), "%s must be a console handle", stream.Name())
	}

	if args[0] != "driver" {
		runInteractiveHelper(t, args)
		return
	}

	require.Len(t, args, 3)
	dir, mode := args[1], args[2]

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	t.Cleanup(func() { signal.Stop(interrupts) })

	switch mode {
	case "cancel":
		checkInteractiveCancellation(t, "TestInteractiveWindowsHelperProcess", dir, interrupts)
	case "ctrl-c":
		checkInteractiveWindowsCtrlC(t, dir, interrupts)
	default:
		require.FailNowf(t, "unknown console driver mode", "%q", mode)
	}

	markInteractiveEvent(t, dir, "driver.done")
}

func checkInteractiveWindowsCtrlC(t *testing.T, dir string, interrupts <-chan os.Signal) {
	t.Helper()

	logger, err := logs.NewNoopLogger("interactive-windows")
	require.NoError(t, err)

	done, pid := startInteractiveHelper(t.Context(), t, logger, "TestInteractiveWindowsHelperProcess", dir, "graceful")

	require.NoError(t, windows.GenerateConsoleCtrlEvent(windows.CTRL_C_EVENT, 0))

	select {
	case <-interrupts:
	case <-time.After(interactiveTestTimeout):
		require.FailNow(t, "driver did not receive the console-wide Ctrl+C positive control")
	}

	assert.Equal(t, pid, waitInteractiveEvent(t, dir, "graceful.interrupted"))
	assert.Equal(t, pid, waitInteractiveEvent(t, dir, "graceful.cleanup"))
	waitInteractiveExit(t, done)
}

func interactiveWindowsAttachConsole(t *testing.T) {
	t.Helper()

	// The executor needs real *os.File handles, not filesystem adapters.
	input, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, input.Close()) })

	output, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, output.Close()) })

	oldIn, oldOut, oldErr := os.Stdin, os.Stdout, os.Stderr
	os.Stdin, os.Stdout, os.Stderr = input, output, output
	t.Cleanup(func() { os.Stdin, os.Stdout, os.Stderr = oldIn, oldOut, oldErr })

	// Clear inherited ignore state only in the driver: doing so in a child would
	// hide a CREATE_NEW_PROCESS_GROUP regression in the interactive executor.
	result, _, callErr := windows.NewLazySystemDLL("kernel32.dll").NewProc("SetConsoleCtrlHandler").Call(0, 0)
	require.NotZero(t, result, "SetConsoleCtrlHandler(NULL, FALSE): %v", callErr)
}
