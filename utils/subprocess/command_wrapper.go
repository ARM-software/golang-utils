/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package subprocess

import (
	"context"
	"os/exec"
	"time"

	"github.com/sasha-s/go-deadlock"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/proc"
	commandUtils "github.com/ARM-software/golang-utils/utils/subprocess/command"
)

// INTERNAL
// wrapper over an exec cmd.
type cmdWrapper struct {
	mu  deadlock.RWMutex
	cmd *exec.Cmd
}

func (c *cmdWrapper) Set(cmd *exec.Cmd) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd == nil {
		c.cmd = cmd
	}
}

func (c *cmdWrapper) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cmd = nil
}

func (c *cmdWrapper) Start() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return commonerrors.UndefinedVariable("command")
	}
	return ConvertCommandError(c.cmd.Start())
}

func (c *cmdWrapper) Run() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return commonerrors.UndefinedVariable("command")
	}
	return ConvertCommandError(c.cmd.Run())
}

func (c *cmdWrapper) interruptWithContext(ctx context.Context, interrupt proc.InterruptType) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return commonerrors.UndefinedVariable("command")
	}
	subprocess := c.cmd.Process
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stopErr := atomic.NewError(nil)
	if subprocess != nil {
		pid := subprocess.Pid
		parallelisation.ScheduleAfter(ctx, proc.SubprocessTerminationGracePeriod, func(time.Time) {
			process, sErr := proc.FindProcess(ctx, pid)
			if process == nil || sErr != nil {
				return
			}
			sErr = proc.InterruptProcess(ctx, pid, interrupt)
			if commonerrors.Any(sErr, commonerrors.ErrInvalid, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
				stopErr.Store(sErr)
			}
		})
	}

	err := parallelisation.WaitWithContextAndError(ctx, c.cmd)
	if commonerrors.Any(err, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
		return err
	}

	return stopErr.Load()
}

func (c *cmdWrapper) interrupt(interrupt proc.InterruptType) error {
	return c.interruptWithContext(context.Background(), interrupt)
}

func (c *cmdWrapper) Stop() error {
	return c.interrupt(proc.SigKill)
}

func (c *cmdWrapper) Interrupt(ctx context.Context) error {
	return c.interruptWithContext(ctx, proc.SigInt)
}

func (c *cmdWrapper) Pid() (pid int, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		err = commonerrors.UndefinedVariable("command")
		return
	}
	subprocess := c.cmd.Process
	if subprocess == nil {
		err = commonerrors.UndefinedVariable("subprocess")
		return
	}
	pid = subprocess.Pid
	return
}

// Definition of a command
type command struct {
	cmd        string
	args       []string
	env        []string
	as         *commandUtils.CommandAsDifferentUser
	loggers    logs.Loggers
	cmdWrapper cmdWrapper
	io         ICommandIO
}

func (c *command) createCommand(cmdCtx context.Context) *exec.Cmd {
	newCmd, newArgs := c.as.Redefine(c.cmd, c.args...)
	cmd := exec.CommandContext(cmdCtx, newCmd, newArgs...) //nolint:gosec
	cancellableCmd, err := proc.DefineCmdCancel(cmd)
	if err == nil {
		cmd = cancellableCmd
	}
	cmd.Stdout = c.io.SetOutput(cmdCtx)
	cmd.Stderr = c.io.SetError(cmdCtx)
	cmd.Stdin = c.io.SetInput(cmdCtx)
	cmd.Env = cmd.Environ()
	cmd.Env = append(cmd.Env, c.env...)
	setGroupAttrToCmd(cmd)
	return cmd
}

func (c *command) GetPath() string {
	return c.cmd
}

func (c *command) GetCmd(cmdCtx context.Context) *cmdWrapper {
	c.cmdWrapper.Set(c.createCommand(cmdCtx))
	return &c.cmdWrapper
}

func (c *command) Reset() {
	c.cmdWrapper.Reset()
}

func (c *command) Check() (err error) {
	if c.cmd == "" {
		err = commonerrors.UndefinedVariable("command")
		return
	}
	if c.as == nil {
		err = commonerrors.UndefinedVariable("command translator")
		return
	}
	if c.loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	return
}

func newCommand(loggers logs.Loggers, as *commandUtils.CommandAsDifferentUser, env []string, cmd string, args ...string) (osCmd *command) {
	osCmd = &command{
		cmd:        cmd,
		args:       args,
		env:        env,
		as:         as,
		loggers:    loggers,
		cmdWrapper: cmdWrapper{},
		io:         NewIOFromLoggers(loggers),
	}
	return
}

func newCommandWithCustomIO(loggers logs.Loggers, io ICommandIO, as *commandUtils.CommandAsDifferentUser, env []string, cmd string, args ...string) (osCmd *command) {
	osCmd = &command{
		cmd:        cmd,
		args:       args,
		env:        env,
		as:         as,
		loggers:    loggers,
		cmdWrapper: cmdWrapper{},
		io:         io,
	}
	return
}

func ConvertCommandError(err error) error {
	return proc.ConvertProcessError(err)
}

// CleanKillOfCommand tries to terminate a command gracefully.
func CleanKillOfCommand(ctx context.Context, cmd *exec.Cmd) (err error) {
	if cmd == nil {
		return
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	thisP := cmd.Process
	if thisP == nil {
		return
	}
	err = proc.TerminateGracefully(ctx, thisP.Pid, proc.SubprocessTerminationGracePeriod)
	return
}
