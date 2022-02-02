/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
// Package subprocess allows you to spawn new processes, retrieve their output/error pipes, and obtain their return codes.
package subprocess

import (
	"context"
	"fmt"

	"github.com/sasha-s/go-deadlock"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"
)

// Subprocess describes what a subproccess is as well as any monitoring it may need.
type Subprocess struct {
	mu                deadlock.RWMutex
	isRunning         atomic.Bool
	command           *command
	processMonitoring *subprocessMonitoring
	messsaging        *subprocessMessaging
}

// New creates a subprocess description.
func New(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.Setup(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	return
}

// Execute executes a command (i.e. spawns a subprocess)
func Execute(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	p, err := New(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

// Setup sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure.
func (s *Subprocess) Setup(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	if s.IsOn() {
		err = s.Stop()
		if err != nil {
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning.Store(false)
	s.processMonitoring = newSubprocessMonitoring(ctx)
	s.command = newCommand(loggers, cmd, args...)
	s.messsaging = newSubprocessMessaging(loggers, messageOnSuccess, messageOnFailure, messageOnStart, s.command.GetPath())
	s.reset()
	return s.check()
}

func (s *Subprocess) check() (err error) {
	// In GO, there is no reentrant locks and so following what is described there
	// https://groups.google.com/forum/#!msg/golang-nuts/XqW1qcuZgKg/Ui3nQkeLV80J
	if s.command == nil {
		err = fmt.Errorf("missing command: %w", commonerrors.ErrUndefined)
		return
	}
	err = s.command.Check()
	if err != nil {
		return
	}
	if s.messsaging == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	err = s.messsaging.Check()
	return
}

// Check checks whether the subprocess is correctly defined.
func (s *Subprocess) Check() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.check()
}

// IsOn states whether the subprocess is running or not.
func (s *Subprocess) IsOn() bool {
	return s.isRunning.Load() && s.processMonitoring.IsOn()
}

// Start starts the process if not already started.
// This method is idempotent.
func (s *Subprocess) Start() (err error) {
	if s.IsOn() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err = s.check()
	if err != nil {
		return
	}
	if s.IsOn() {
		return
	}
	s.reset()
	s.runProcessMonitoring()
	cmd := s.getCmd()
	err = cmd.Start()
	if err != nil {
		s.messsaging.LogFailedStart(err)
		s.isRunning.Store(false)
		s.Cancel()
		return
	}
	pid, err := cmd.Pid()
	if err != nil {
		s.messsaging.LogFailedStart(err)
		s.isRunning.Store(false)
		s.Cancel()
		return
	}

	s.isRunning.Store(true)
	s.messsaging.SetPid(pid)
	s.messsaging.LogStarted()
	return
}

// Cancel interrupts an on-going process. This method is idempotent.
func (s *Subprocess) Cancel() {
	s.processMonitoring.CancelSubprocess()
}

// Execute executes the command and waits for completion.
func (s *Subprocess) Execute() (err error) {
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.Cancel()

	if s.IsOn() {
		return fmt.Errorf("process is already started: %w", commonerrors.ErrConflict)
	}
	s.processMonitoring.Reset()
	s.command.Reset()
	s.messsaging.LogStart()
	s.runProcessMonitoring()
	cmd := s.getCmd()
	s.isRunning.Store(true)
	err = cmd.Run()
	s.isRunning.Store(false)
	s.messsaging.LogEnd(err)
	return
}

// Stop stops the process straight away if currently working without waiting for completion. This method should be used in combination with `Start`.
// However, in order to interrupt a process however it was started (using `Start` or `Execute`), prefer `Cancel`.
// This method is idempotent.
func (s *Subprocess) Stop() (err error) {
	return s.stop(true)
}

// Restart restarts a process. It will stop the process if currently running.
func (s *Subprocess) Restart() (err error) {
	err = s.stop(false)
	if err != nil {
		return
	}
	return s.Start()
}
func (s *Subprocess) getCmd() *cmdWrapper {
	return s.command.GetCmd(s.processMonitoring.ProcessContext())
}

func (s *Subprocess) runProcessMonitoring() {
	s.processMonitoring.RunMonitoring(s.Stop)
}

func (s *Subprocess) reset() {
	s.processMonitoring.Reset()
	s.command.Reset()
}

func (s *Subprocess) stop(cancel bool) (err error) {
	if !s.IsOn() {
		return
	}
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if cancel {
			s.Cancel()
		}
	}()
	if !s.IsOn() {
		return
	}
	s.messsaging.LogStopping()
	err = s.getCmd().Stop()
	s.command.Reset()
	s.isRunning.Store(false)
	s.messsaging.LogEnd(nil)
	return
}
