/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"context"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
	"github.com/ARM-software/golang-utils/utils/platform"
	commandUtils "github.com/ARM-software/golang-utils/utils/subprocess/command"
)

func TestCmdRun(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
	}{
		{
			name:       "ShortProcess",
			cmdWindows: "cmd",
			argWindows: []string{"dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"SLEEP 1"},
			cmdOther:   "sleep",
			argOther:   []string{"1"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			var cmd *command
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.NoError(t, err)
			if platform.IsWindows() {
				cmd = newCommand(loggers, commandUtils.Me(), nil, test.cmdWindows, test.argWindows...)
			} else {
				cmd = newCommand(loggers, commandUtils.Me(), nil, test.cmdOther, test.argOther...)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			wrapper := cmd.GetCmd(ctx)
			err = wrapper.Run()
			require.NoError(t, err)
			err = wrapper.Run()
			require.Error(t, err)
			cmd.Reset()
			wrapper = cmd.GetCmd(ctx)
			err = wrapper.Run()
			require.NoError(t, err)
		})
	}
}

func TestCmdRunWithEnv(t *testing.T) {
	envTest := struct {
		cmdWindows string
		cmdOther   string
		envVars    []string
	}{
		cmdWindows: "Env",
		cmdOther:   "env",
		envVars:    []string{"TEST1=TEST2", "TEST3=TEST4"},
	}

	t.Run("Test command run with env vars", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		var cmd *command
		loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
		require.NoError(t, err)
		if platform.IsWindows() {
			cmd = newCommand(loggers, commandUtils.Me(), envTest.envVars, "powershell", "-Command", envTest.cmdWindows)
		} else {
			cmd = newCommand(loggers, commandUtils.Me(), envTest.envVars, envTest.cmdOther)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		wrapper := cmd.GetCmd(ctx)
		wrapper.cmd.Stdout = nil
		out, err := wrapper.cmd.Output()
		require.NoError(t, err)

		for i := range envTest.envVars {
			envVar := envTest.envVars[i]
			assert.True(t, regexp.MustCompile(envVar).Match(out))
		}
	})
}

func TestCmdStartStop(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
	}{
		{
			name:       "ShortProcess",
			cmdWindows: "cmd",
			argWindows: []string{"dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"SLEEP 4"},
			cmdOther:   "sleep",
			argOther:   []string{"4"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			var cmd *command
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.NoError(t, err)

			if platform.IsWindows() {
				cmd = newCommand(loggers, commandUtils.Me(), nil, test.cmdWindows, test.argWindows...)
			} else {
				cmd = newCommand(loggers, commandUtils.Me(), nil, test.cmdOther, test.argOther...)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			wrapper := cmd.GetCmd(ctx)
			err = wrapper.Start()
			require.NoError(t, err)
			pid, err := wrapper.Pid()
			require.NoError(t, err)
			assert.NotZero(t, pid)
			err = wrapper.Start()
			require.Error(t, err)
			err = wrapper.Stop()
			require.NoError(t, err)
			err = wrapper.Start()
			require.Error(t, err)
			err = wrapper.Interrupt()
			require.NoError(t, err)
		})
	}
}
