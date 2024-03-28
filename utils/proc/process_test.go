/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package proc

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestFindProcess(t *testing.T) {
	p, err := FindProcess(context.Background(), os.Getpid())
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, os.Getpid(), p.Pid())
}

func TestProcesses(t *testing.T) {
	// This test works because there will always be SOME processes
	// running.
	p, err := Ps(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, p)

	var executables []string
	var names []string
	var cmdLines []string
	var cwds []string
	for i := range p {
		ex := strings.TrimSpace(strings.TrimSuffix(filepath.Base(p[i].Executable()), ".exe"))
		if ex != "" {
			executables = append(executables, ex)
		}
		name := strings.TrimSuffix(strings.TrimSpace(p[i].Name()), ".exe")
		if name != "" {
			names = append(names, name)
		}
		cwd := strings.TrimSpace(p[i].Cwd())
		if cwd != "" {
			cwds = append(cwds, cwd)
		}
		cmdLine := strings.TrimSpace(p[i].Cmdline())
		if cmdLine != "" {
			cmdLines = append(cmdLines, cmdLine)
		}
	}
	assert.Contains(t, executables, "go")
	assert.NotEmpty(t, cwds)
	assert.NotEmpty(t, cmdLines)
	assert.Contains(t, names, "go")
}

func TestKill(t *testing.T) {
	cmd := exec.Command("sleep", "50")
	require.NoError(t, cmd.Start())
	defer func() { _ = cmd.Wait() }()
	process, err := FindProcess(context.Background(), cmd.Process.Pid)
	require.NoError(t, err)
	assert.True(t, process.IsRunning())
	require.NoError(t, process.Terminate(context.Background()))
	require.NoError(t, process.KillWithChildren(context.Background()))
	process, err = FindProcess(context.Background(), cmd.Process.Pid)
	if err == nil {
		require.NotEmpty(t, process)
		assert.False(t, process.IsRunning())
	} else {
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, process)
	}
}

func TestPs_KillWithChildren(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test with bash")
	}
	defer goleak.VerifyNone(t)
	// see https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	// https://forum.golangbridge.org/t/killing-child-process-on-timeout-in-go-code/995/16
	cmd := exec.Command("bash", "-c", "watch date > date.txt 2>&1")
	require.NoError(t, cmd.Start())
	defer func() { _ = cmd.Wait() }()
	require.NotNil(t, cmd.Process)
	p, err := FindProcess(context.Background(), cmd.Process.Pid)
	require.NoError(t, err)
	assert.True(t, p.IsRunning())
	require.NoError(t, p.KillWithChildren(context.Background()))
	p, err = FindProcess(context.Background(), cmd.Process.Pid)
	if err == nil {
		require.NotEmpty(t, p)
		assert.False(t, p.IsRunning())
	} else {
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, p)
	}
}
