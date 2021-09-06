/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
//The subprocess module allows you to spawn new processes, retrieve their output/error pipes, and obtain their return codes.
package subprocess

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// A subprocess description.
type Subprocess struct {
	mu             sync.RWMutex
	parentCtx      context.Context
	cancellableCtx atomic.Value
	cancelStore    *parallelisation.CancelFunctionStore
	cmdCanceller   context.CancelFunc
	command        *exec.Cmd
	subprocess     *os.Process
	isRunning      atomic.Bool
	messsaging     *subprocessMessaging
}

// Creates a subprocess description.
func New(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.Setup(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	return
}

// Executes a command (i.e. spawns a subprocess)
func Execute(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	p, err := New(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

func (s *Subprocess) check() (err error) {
	// In GO, there is no reentrant locks and so following what is described there
	// https://groups.google.com/forum/#!msg/golang-nuts/XqW1qcuZgKg/Ui3nQkeLV80J
	if s.command == nil {
		err = fmt.Errorf("missing command: %w", commonerrors.ErrUndefined)
		return
	}
	if s.messsaging == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	err = s.messsaging.Check()
	return
}

// Checks whether the subprocess is correctly defined.
func (s *Subprocess) Check() (err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.check()
}

func (s *Subprocess) resetContext() {
	subctx, cancelFunc := context.WithCancel(s.parentCtx)
	s.cancellableCtx.Store(subctx)
	s.cancelStore.RegisterCancelFunction(cancelFunc)
}

// Sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure.
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
	s.cancelStore = parallelisation.NewCancelFunctionsStore()
	s.parentCtx = ctx
	cmdCtx, cmdcancelFunc := context.WithCancel(s.parentCtx)
	s.cmdCanceller = cmdcancelFunc
	s.command = exec.CommandContext(cmdCtx, cmd, args...)
	s.command.Stdout = newOutStreamer(loggers)
	s.command.Stderr = newErrLogStreamer(loggers)
	s.messsaging = NewSubprocessMessaging(loggers, messageOnSuccess, messageOnFailure, messageOnStart, s.command.Path)
	return s.check()
}

// States whether the subprocess is on or not.
func (s *Subprocess) IsOn() bool {
	return s.isRunning.Load()
}

// Starts the process if not already started.
func (s *Subprocess) Start() (err error) {
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.IsOn() {
		return
	}
	s.runProcessStatusCheck()
	err = commonerrors.ConvertContextError(s.command.Start())
	if err != nil {
		s.messsaging.LogFailedStart(err)
		s.isRunning.Store(false)
		return
	}
	s.subprocess = s.command.Process
	s.isRunning.Store(true)
	s.messsaging.SetPid(s.subprocess.Pid)
	s.messsaging.LogStarted()
	return
}

func (s *Subprocess) Cancel() {
	store := s.cancelStore
	if store != nil {
		store.Cancel()
	}
}

func (s *Subprocess) runProcessStatusCheck() {
	s.resetContext()
	go func(proc *Subprocess) {
		<-proc.cancellableCtx.Load().(context.Context).Done()
		proc.Cancel()
		_ = proc.Stop()
	}(s)
}

// Executes the command and waits for completion.
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
	s.messsaging.LogStart()
	s.cancelStore.RegisterCancelFunction(s.cmdCanceller)
	s.runProcessStatusCheck()
	s.isRunning.Store(true)
	err = commonerrors.ConvertContextError(s.command.Run())
	s.isRunning.Store(false)
	s.messsaging.LogEnd(err)
	return
}

// Stops the process if currently working.
func (s *Subprocess) Stop() (err error) {
	if !s.IsOn() {
		return
	}
	err = s.Check()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.Cancel()
	if !s.IsOn() {
		return
	}
	s.messsaging.LogStopping()
	_ = s.subprocess.Kill()
	_ = s.command.Wait()
	s.command.Process = nil
	s.subprocess = nil
	s.isRunning.Store(false)
	s.messsaging.LogEnd(nil)
	return
}

// Restarts a process.
func (s *Subprocess) Restart() (err error) {
	err = s.Stop()
	if err != nil {
		return
	}
	return s.Start()
}
