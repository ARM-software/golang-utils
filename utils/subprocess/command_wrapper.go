/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package subprocess

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/sasha-s/go-deadlock"

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
		return fmt.Errorf("%w:undefined command", commonerrors.ErrUndefined)
	}
	return ConvertCommandError(c.cmd.Start())
}

func (c *cmdWrapper) Run() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return fmt.Errorf("%w:undefined command", commonerrors.ErrUndefined)
	}
	return ConvertCommandError(c.cmd.Run())
}

type interruptType int

const (
	sigint  interruptType = 2
	sigkill interruptType = 9
	sigterm interruptType = 15
)

func (c *cmdWrapper) interruptWithContext(ctx context.Context, interrupt interruptType) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		return commonerrors.New(commonerrors.ErrUndefined, "undefined command")
	}
	subprocess := c.cmd.Process
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var stopErr error
	if subprocess != nil {
		pid := subprocess.Pid
		parallelisation.ScheduleAfter(ctx, 10*time.Millisecond, func(time.Time) {
			process, err := proc.FindProcess(ctx, pid)
			if process == nil || err != nil {
				return
			}
			switch interrupt {
			case sigint:
				_ = process.Interrupt(ctx)
			case sigkill:
				_ = process.KillWithChildren(ctx)
			case sigterm:
				_ = process.Terminate(ctx)
			default:
				stopErr = commonerrors.New(commonerrors.ErrInvalid, "unknown interrupt type for process")
			}
		})
	}

	err := parallelisation.WaitWithContextAndError(ctx, c.cmd)
	if commonerrors.Any(err, commonerrors.ErrCancelled, commonerrors.ErrTimeout) {
		return err
	}

	return stopErr
}

func (c *cmdWrapper) interrupt(interrupt interruptType) error {
	return c.interruptWithContext(context.Background(), interrupt)
}

func (c *cmdWrapper) Stop() error {
	return c.interrupt(sigkill)
}

func (c *cmdWrapper) Interrupt(ctx context.Context) error {
	return c.interruptWithContext(ctx, sigint)
}

func (c *cmdWrapper) Pid() (pid int, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cmd == nil {
		err = fmt.Errorf("%w:undefined command", commonerrors.ErrUndefined)
		return
	}
	subprocess := c.cmd.Process
	if subprocess == nil {
		err = fmt.Errorf("%w:undefined subprocess", commonerrors.ErrUndefined)
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
}

func (c *command) createCommand(cmdCtx context.Context) *exec.Cmd {
	newCmd, newArgs := c.as.Redefine(c.cmd, c.args...)
	cmd := exec.CommandContext(cmdCtx, newCmd, newArgs...) //nolint:gosec
	cmd.Stdout = newOutStreamer(cmdCtx, c.loggers)
	cmd.Stderr = newErrLogStreamer(cmdCtx, c.loggers)
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
		err = fmt.Errorf("missing command: %w", commonerrors.ErrUndefined)
		return
	}
	if c.as == nil {
		err = fmt.Errorf("missing command translator: %w", commonerrors.ErrUndefined)
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
	}
	return
}

func ConvertCommandError(err error) error {
	return proc.ConvertProcessError(err)
}

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
	} else {
		p, subErr := proc.FindProcess(ctx, thisP.Pid)
		if subErr != nil {
			err = subErr
			return
		}
		err = p.KillWithChildren(ctx)
	}
	return
}
