/*
 * Copyright (C) 2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs"
)

func TestInteractiveCommandPolicy(t *testing.T) {
	for _, interactive := range []bool{false, true} {
		t.Run(fmt.Sprintf("interactive=%t", interactive), func(t *testing.T) {
			logger, err := logs.NewPlainStringLogger()
			require.NoError(t, err)

			process, err := NewWithIOWithDir(t.Context(), logger, NewDefaultIO(), "", "", "", "", os.Args[0])
			require.NoError(t, err)
			t.Cleanup(process.Cancel)
			assert.False(t, process.command.interactive, "existing constructors must retain batch defaults")

			process.command.interactive = interactive
			command := process.command.createCommand(t.Context())

			assert.Same(t, os.Stdin, command.Stdin)
			assert.Same(t, os.Stdout, command.Stdout)
			assert.Same(t, os.Stderr, command.Stderr)
			assert.NotNil(t, command.Cancel)

			if interactive {
				assert.Nil(t, command.SysProcAttr, "interactive commands must inherit process-group membership")
				assert.Equal(t, interactiveTerminationGracePeriod, command.WaitDelay)
			} else {
				assert.NotNil(t, command.SysProcAttr, "batch commands must retain process-group isolation")
				assert.Zero(t, command.WaitDelay)
			}
		})
	}
}

func TestExecuteInteractiveWithDirIO(t *testing.T) {
	dir := t.TempDir()
	wantArgs := []string{"two words", "", `quote"and'`, "$HOME; literal", `trailing\`, "--flag=value"}
	command, args := interactiveHelperCommand(t, interactiveHelperTest, "report", dir, wantArgs...)
	t.Setenv("INTERACTIVE_TEST_VALUE", "inherited value=with spaces")

	stdin, stdout, stderr := interactiveStreamFile(t), interactiveStreamFile(t), interactiveStreamFile(t)
	_, err := stdin.WriteString("input without a final newline")
	require.NoError(t, err)
	_, err = stdin.Seek(0, io.SeekStart)
	require.NoError(t, err)

	oldIn, oldOut, oldErr := os.Stdin, os.Stdout, os.Stderr
	os.Stdin, os.Stdout, os.Stderr = stdin, stdout, stderr
	t.Cleanup(func() { os.Stdin, os.Stdout, os.Stderr = oldIn, oldOut, oldErr })

	logger, err := logs.NewPlainStringLogger()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), interactiveTestTimeout)
	defer cancel()

	require.NoError(t, ExecuteInteractiveWithDir(ctx, logger, "start", "success", "failure", dir, command, args...))

	output, err := filesystem.ReadFile(stdout.Name())
	require.NoError(t, err)
	var report []string
	require.NoError(t, json.Unmarshal(output, &report))
	require.Len(t, report, 3+len(wantArgs))

	wantDir, err := filesystem.Stat(dir)
	require.NoError(t, err)
	gotDir, err := filesystem.Stat(report[0])
	require.NoError(t, err)
	assert.True(t, os.SameFile(wantDir, gotDir))
	assert.Equal(t, append([]string{"inherited value=with spaces", "input without a final newline"}, wantArgs...), report[1:])

	output, err = filesystem.ReadFile(stderr.Name())
	require.NoError(t, err)
	assert.Equal(t, "child stderr\n", string(output))
	assert.Equal(t, "start\nsuccess\n", logger.GetLogContent())
}

func TestExecuteInteractiveWithDirLifecycle(t *testing.T) {
	for _, test := range []struct {
		name, role string
		custom     bool
	}{
		{"default success", "success", false},
		{"default failure", "failure", false},
		{"custom success", "success", true},
		{"custom failure", "failure", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			command, args := interactiveHelperCommand(t, interactiveHelperTest, test.role, t.TempDir())
			logger, err := logs.NewPlainStringLogger()
			require.NoError(t, err)

			start, success, failure := "", "", ""
			if test.custom {
				start, success, failure = "custom start", "custom success", "custom failure"
			}

			ctx, cancel := context.WithTimeout(t.Context(), interactiveTestTimeout)
			defer cancel()

			err = ExecuteInteractiveWithDir(ctx, logger, start, success, failure, "", command, args...)

			if test.role == "failure" {
				var exitError *exec.ExitError
				require.ErrorAs(t, err, &exitError)
				assert.Equal(t, 23, exitError.ExitCode())
			} else {
				require.NoError(t, err)
			}

			if !test.custom {
				start = fmt.Sprintf("Executing command  -> `%v`", command)
				success = fmt.Sprintf("command  -> `%v` ended successfully", command)
				failure = fmt.Sprintf("Error occurred when executing -> `%v`: ", command)
			}

			lines := strings.Split(strings.TrimSuffix(logger.GetLogContent(), "\n"), "\n")
			require.Len(t, lines, 2)
			assert.Equal(t, start, lines[0])

			if test.role == "failure" {
				assert.True(t, strings.HasPrefix(lines[1], failure), lines[1])
				assert.NotContains(t, lines[1], success)
			} else {
				assert.Equal(t, success, lines[1])
			}
		})
	}
}

func TestExecuteInteractiveWithDirInvalid(t *testing.T) {
	for _, test := range []struct {
		name string
		want error
	}{
		{"empty command", commonerrors.ErrUndefined},
		{"nil logger", commonerrors.ErrNoLogger},
		{"invalid logger", commonerrors.ErrNoLogger},
		{"cancelled", commonerrors.ErrCancelled},
		{"expired", commonerrors.ErrTimeout},
		{"missing command", os.ErrNotExist},
		{"missing directory", os.ErrNotExist},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			command, args := interactiveHelperCommand(t, interactiveHelperTest, "success", dir)
			plainLogger, err := logs.NewPlainStringLogger()
			require.NoError(t, err)
			var logger logs.Loggers = plainLogger

			ctx, cancel := context.WithTimeout(t.Context(), interactiveTestTimeout)
			defer cancel()
			workingDir := dir

			switch test.name {
			case "empty command":
				command = ""
			case "nil logger":
				logger = nil
			case "invalid logger":
				logger = &logs.GenericLoggers{}
			case "cancelled":
				cancel()
			case "expired":
				expired, stop := context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer stop()
				ctx = expired
			case "missing command":
				command = filepath.Join(dir, "missing-executable")
			case "missing directory":
				workingDir = filepath.Join(dir, "missing-directory")
			}

			err = ExecuteInteractiveWithDir(ctx, logger, "start", "success", "failure", workingDir, command, args...)

			errortest.RequireError(t, err, test.want)
			_, err = filesystem.Stat(filepath.Join(dir, "started"))
			errortest.RequireError(t, err, commonerrors.ErrNotFound)

			if test.name != "missing command" && test.name != "missing directory" {
				assert.Empty(t, plainLogger.GetLogContent())
			}
		})
	}
}

func TestExecuteInteractiveWithDirCancellation(t *testing.T) {
	for _, test := range []struct {
		name, role string
		deadline   bool
	}{
		{"graceful", "graceful", false},
		{"forced", "ignore", false},
		{"deadline", "graceful", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && test.name == "graceful" {
				t.Skip("Windows does not support signalling os.Interrupt to a child")
			}

			dir := t.TempDir()
			logger, err := logs.NewPlainStringLogger()
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			done, pid := startInteractiveHelper(ctx, t, logger, interactiveHelperTest, dir, test.role)

			stopAt := time.Now()
			want := commonerrors.ErrCancelled
			if test.deadline {
				stopAt, _ = ctx.Deadline()
				want = commonerrors.ErrTimeout
			} else {
				cancel()
			}

			waitInteractiveExit(t, done, want)
			assert.Contains(t, logger.GetLogContent(), "failure")
			assert.NotContains(t, logger.GetLogContent(), "success")

			if runtime.GOOS != "windows" {
				assert.Equal(t, pid, waitInteractiveEvent(t, dir, fmt.Sprintf("%s.interrupted", test.role)))

				if test.role == "ignore" {
					assert.GreaterOrEqual(t, time.Since(stopAt), interactiveTerminationGracePeriod)
				} else {
					assert.Equal(t, pid, waitInteractiveEvent(t, dir, "graceful.cleanup"))
				}
			}
		})
	}
}

func TestInteractiveHelperProcess(t *testing.T) {
	args := interactiveHelperArguments(t)
	if args == nil {
		return
	}

	runInteractiveHelper(t, args)
	os.Exit(0)
}
