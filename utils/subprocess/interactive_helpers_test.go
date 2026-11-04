/*
 * Copyright (C) 2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// Helpers re-execute the test binary. PID marker files coordinate readiness and
// completion independently of the terminal used by each platform fixture.
const (
	interactiveHelperGuard   = "GO_SUBPROCESS_INTERACTIVE_HELPER"
	interactiveHelperTest    = "TestInteractiveHelperProcess"
	interactiveTestTimeout   = 10 * time.Second
	interactiveHelperTimeout = 30 * time.Second
)

func interactiveHelperCommand(t *testing.T, helperTest, role, dir string, args ...string) (executable string, arguments []string) {
	t.Helper()

	t.Setenv(interactiveHelperGuard, "1")
	executable, err := os.Executable()
	require.NoError(t, err)

	arguments = append([]string{fmt.Sprintf("-test.run=^%s$", helperTest), "--", role, dir}, args...)
	return
}

func interactiveHelperArguments(t *testing.T) (args []string) {
	t.Helper()

	if os.Getenv(interactiveHelperGuard) != "1" {
		return
	}

	watchdog := time.AfterFunc(interactiveHelperTimeout, func() { os.Exit(99) })
	t.Cleanup(func() { watchdog.Stop() })

	separator := slices.Index(os.Args, "--")
	require.NotEqual(t, -1, separator)

	args = os.Args[separator+1:]
	require.GreaterOrEqual(t, len(args), 2)

	return
}

func markInteractiveEvent(t *testing.T, dir, name string) {
	t.Helper()

	require.NoError(t, filesystem.WriteFile(filepath.Join(dir, name), []byte(fmt.Sprintf("%d\n", os.Getpid())), 0o600))
}

func waitInteractiveEvent(t *testing.T, dir, name string) (pid int) {
	t.Helper()

	require.Eventually(t, func() bool {
		data, err := filesystem.ReadFile(filepath.Join(dir, name))
		if err != nil || !strings.HasSuffix(string(data), "\n") {
			return false
		}

		pid, err = strconv.Atoi(strings.TrimSuffix(string(data), "\n"))
		return err == nil && pid > 0
	}, interactiveTestTimeout, 10*time.Millisecond, "helper did not report %s", name)

	return
}

func waitInteractiveExit(t *testing.T, done <-chan error, expectedErrors ...error) {
	t.Helper()

	var err error
	select {
	case err = <-done:
	case <-time.After(interactiveTestTimeout):
		require.FailNow(t, "helper did not exit before the deadline")
	}

	if len(expectedErrors) == 0 {
		require.NoError(t, err)
		return
	}

	errortest.RequireError(t, err, expectedErrors...)
}

func startInteractiveTestProcess(t *testing.T, command *exec.Cmd) <-chan error {
	t.Helper()

	command.WaitDelay = time.Second
	require.NoError(t, command.Start())

	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
		close(done)
	}()

	t.Cleanup(func() {
		select {
		case <-done:
			return
		default:
			_ = command.Process.Kill()
		}

		select {
		case <-done:
		case <-time.After(interactiveTestTimeout):
			assert.Fail(t, "helper did not exit during cleanup")
		}
	})

	return done
}

func startInteractiveHelper(ctx context.Context, t *testing.T, logger logs.Loggers, helperTest, dir, role string) (<-chan error, int) {
	t.Helper()

	require.NoError(t, parallelisation.DetermineContextError(ctx))

	ctx, cancel := context.WithCancel(ctx)
	command, args := interactiveHelperCommand(t, helperTest, role, dir)

	done := make(chan error, 1)
	var child *os.Process

	t.Cleanup(func() {
		cancel()

		if child != nil {
			defer func() { _ = child.Release() }()
		}

		select {
		case <-done:
			return
		default:
			if child != nil {
				_ = child.Kill()
			}
		}

		// If startup failed before reporting a PID, the helper watchdog still bounds cleanup.
		select {
		case <-done:
		case <-time.After(interactiveHelperTimeout + time.Second):
			assert.Fail(t, "interactive executor did not reap its helper")
		}
	})

	go func() {
		done <- ExecuteInteractiveWithDir(ctx, logger, "start", "success", "failure", dir, command, args...)
		close(done)
	}()

	pid := waitInteractiveEvent(t, dir, fmt.Sprintf("%s.ready", role))
	child, err := os.FindProcess(pid)
	require.NoError(t, err)

	return done, pid
}

func checkInteractiveCancellation(t *testing.T, helperTest, dir string, interrupts <-chan os.Signal) {
	t.Helper()

	command, args := interactiveHelperCommand(t, helperTest, "sibling", dir)
	sibling := exec.CommandContext(t.Context(), command, args...)
	sibling.Stdin, sibling.Stdout, sibling.Stderr = os.Stdin, os.Stdout, os.Stderr

	siblingDone := startInteractiveTestProcess(t, sibling)
	assert.Equal(t, sibling.Process.Pid, waitInteractiveEvent(t, dir, "sibling.ready"))

	logger, err := logs.NewNoopLogger("interactive-cancellation")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(t.Context(), interactiveTestTimeout)
	defer cancel()
	childDone, pid := startInteractiveHelper(ctx, t, logger, helperTest, dir, "graceful")

	cancel()
	waitInteractiveExit(t, childDone, commonerrors.ErrCancelled)

	if runtime.GOOS == "windows" {
		assert.False(t, filesystem.Exists(filepath.Join(dir, "graceful.interrupted")), "Windows cancellation must not broadcast an interrupt")
	} else {
		assert.Equal(t, pid, waitInteractiveEvent(t, dir, "graceful.cleanup"))
	}

	// A fresh reply proves the sibling is still running, not merely present in a process list.
	markInteractiveEvent(t, dir, "ping")
	assert.Equal(t, sibling.Process.Pid, waitInteractiveEvent(t, dir, "sibling.ack"))

	markInteractiveEvent(t, dir, "stop")
	waitInteractiveExit(t, siblingDone)
	assert.False(t, filesystem.Exists(filepath.Join(dir, "sibling.interrupted")))

	select {
	case <-interrupts:
		assert.Fail(t, "cancelling the child interrupted its caller")
	case <-time.After(200 * time.Millisecond):
	}
}

func interactiveStreamFile(t *testing.T) (file *os.File) {
	t.Helper()

	// Standard streams require real *os.File descriptors, not filesystem adapters.
	file, err := os.CreateTemp(t.TempDir(), "stream")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })

	return
}

func runInteractiveHelper(t *testing.T, args []string) {
	t.Helper()

	role, dir := args[0], args[1]
	markInteractiveEvent(t, dir, "started")

	switch role {
	case "report":
		input, err := io.ReadAll(os.Stdin)
		require.NoError(t, err)

		cwd, err := filesystem.CurrentDirectory()
		require.NoError(t, err)

		require.NoError(t, json.NewEncoder(os.Stdout).Encode(append([]string{cwd, os.Getenv("INTERACTIVE_TEST_VALUE"), string(input)}, args[2:]...)))

		_, err = fmt.Fprintln(os.Stderr, "child stderr")
		require.NoError(t, err)
	case "failure":
		os.Exit(23)
	case "success":
	case "read":
		markInteractiveEvent(t, dir, "read.ready")

		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		require.NoError(t, err)
		require.NoError(t, filesystem.WriteFile(filepath.Join(dir, "read.input"), []byte(line), 0o600))
	case "graceful", "ignore":
		waitInteractiveInterrupt(t, dir, role)
	case "sibling":
		runInteractiveSibling(t, dir)
	default:
		require.FailNowf(t, "unknown helper role", "%q", role)
	}
}

func waitInteractiveInterrupt(t *testing.T, dir, role string) {
	t.Helper()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)

	markInteractiveEvent(t, dir, fmt.Sprintf("%s.ready", role))

	for range interrupts {
		markInteractiveEvent(t, dir, fmt.Sprintf("%s.interrupted", role))

		if role == "graceful" {
			markInteractiveEvent(t, dir, "graceful.cleanup")
			return
		}
	}
}

func runInteractiveSibling(t *testing.T, dir string) {
	t.Helper()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)

	markInteractiveEvent(t, dir, "sibling.ready")

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	replied := false

	for {
		select {
		case <-interrupts:
			markInteractiveEvent(t, dir, "sibling.interrupted")
		case <-ticker.C:
			if !replied && filesystem.Exists(filepath.Join(dir, "ping")) {
				markInteractiveEvent(t, dir, "sibling.ack")
				replied = true
			}

			if filesystem.Exists(filepath.Join(dir, "stop")) {
				require.Empty(t, interrupts, "sibling has a pending interrupt")
				return
			}
		}
	}
}
