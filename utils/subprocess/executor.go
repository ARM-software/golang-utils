/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package subprocess allows you to spawn new processes, log their output/error and obtain their return codes.
package subprocess

import (
	"context"

	"github.com/sasha-s/go-deadlock"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/proc"
	commandUtils "github.com/ARM-software/golang-utils/utils/subprocess/command"
)

// Subprocess describes what a subprocess is as well as any monitoring it may need.
type Subprocess struct {
	mu                deadlock.RWMutex
	isRunning         atomic.Bool
	command           *command
	processMonitoring *subprocessMonitoring
	messaging         *subprocessMessaging
}

// New creates a subprocess description.
func New(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (*Subprocess, error) {
	return NewWithEnvironment(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
}

// NewWithEnvironment creates a subprocess description. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func NewWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	p, err = newSubProcess(ctx, loggers, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Me(), cmd, args...)
	return
}

// newSubProcess creates a subprocess description.
func newSubProcess(ctx context.Context, loggers logs.Loggers, env []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.SetupAsWithEnvironmentWithCustomIO(ctx, loggers, io, env, messageOnStart, messageOnSuccess, messageOnFailure, as, cmd, args...)
	return
}

func newPlainSubProcess(ctx context.Context, loggers logs.Loggers, env []string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.setup(ctx, loggers, nil, env, false, "", "", "", as, cmd, args...)
	return
}

// ExecuteWithEnvironment executes a command (i.e. spawns a subprocess). It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	p, err := NewWithEnvironment(ctx, loggers, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

// Execute executes a command (i.e. spawns a subprocess).
func Execute(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) error {
	return ExecuteWithEnvironment(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
}

// ExecuteAs executes a command (i.e. spawns a subprocess) as a different user.
func ExecuteAs(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) error {
	return ExecuteAsWithEnvironment(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, as, cmd, args...)
}

// ExecuteAsWithEnvironment executes a command (i.e. spawns a subprocess) as a different user. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteAsWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	p, err := newSubProcess(ctx, loggers, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, as, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

// ExecuteWithSudo executes a command (i.e. spawns a subprocess) as root.
func ExecuteWithSudo(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) error {
	return ExecuteAs(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Sudo(), cmd, args...)
}

// Output executes a command and returns its output (stdOutput and stdErr are merged) as string.
func Output(ctx context.Context, loggers logs.Loggers, cmd string, args ...string) (string, error) {
	return OutputWithEnvironment(ctx, loggers, nil, cmd, args...)
}

// OutputWithEnvironment executes a command and returns its output (stdOutput and stdErr are merged) as string. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func OutputWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, cmd string, args ...string) (string, error) {
	return OutputAsWithEnvironment(ctx, loggers, additionalEnvVars, commandUtils.Me(), cmd, args...)
}

// OutputAs executes a command as a different user and returns its output (stdOutput and stdErr are merged) as string.
func OutputAs(ctx context.Context, loggers logs.Loggers, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (string, error) {
	return OutputAsWithEnvironment(ctx, loggers, nil, as, cmd, args...)
}

// OutputAsWithEnvironment executes a command as a different user and returns its output (stdOutput and stdErr are merged) as string. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func OutputAsWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (output string, err error) {
	if loggers == nil {
		err = commonerrors.ErrNoLogger
		return
	}

	stringLogger, err := logs.NewPlainStringLogger()
	if err != nil {
		return
	}
	mLoggers, err := logs.NewCombinedLoggers(loggers, stringLogger)
	if err != nil {
		return
	}
	p, err := newPlainSubProcess(ctx, mLoggers, additionalEnvVars, as, cmd, args...)
	if err != nil {
		return
	}
	err = p.Execute()
	output = stringLogger.GetLogContent()
	return
}

// Setup sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure.
func (s *Subprocess) Setup(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.SetupWithEnvironment(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
}

// SetupWithEnvironment sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure. Compared to Setup, it allows specifying additional environment variables to be used by the process.
func (s *Subprocess) SetupWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, nil, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Me(), cmd, args...)
}

// SetupAs is similar to Setup but allows the command to be run as a different user.
func (s *Subprocess) SetupAs(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.SetupAsWithEnvironment(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, as, cmd, args...)
}

// SetupAsWithEnvironment is similar to Setup but allows the command to be run as a different user. Compared to SetupAs, it allows specifying additional environment variables to be used by the process.
func (s *Subprocess) SetupAsWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, nil, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, as, cmd, args...)
}

// SetupWithCustomIO sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.SetupWithEnvironmentWithCustomIO(ctx, loggers, io, nil, messageOnStart, messageOnSuccess, messageOnFailure, cmd, args...)
}

// SetupWithEnvironmentWithCustomIO sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure. Compared to SetupWithCustomIO, it allows specifying additional environment variables to be used by the process. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupWithEnvironmentWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, io, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Me(), cmd, args...)
}

// SetupAsWithCustomIO is similar to SetupWithCustomIO but allows the command to be run as a different user. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupAsWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.SetupAsWithEnvironmentWithCustomIO(ctx, loggers, io, nil, messageOnStart, messageOnSuccess, messageOnFailure, as, cmd, args...)
}

// SetupAsWithEnvironmentWithCustomIO is similar to SetupWithCustomIO but allows the command to be run as a different user. Compared to SetupAsWithCustomIO, it allows specifying additional environment variables to be used by the process. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupAsWithEnvironmentWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, io, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, as, cmd, args...)
}

// Setup sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure as well as the stdin, stdout, and stderr.
func (s *Subprocess) setup(ctx context.Context, loggers logs.Loggers, io ICommandIO, env []string, withAdditionalMessages bool, messageOnStart, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
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
	if io != nil {
		s.command = newCommandWithCustomIO(loggers, io, as, env, cmd, args...)
	} else {
		s.command = newCommand(loggers, as, env, cmd, args...)
	}
	s.messaging = newSubprocessMessaging(loggers, withAdditionalMessages, messageOnSuccess, messageOnFailure, messageOnStart, s.command.GetPath())
	s.reset()
	return s.check()
}

func (s *Subprocess) check() (err error) {
	// In GO, there is no reentrant locks and so following what is described there
	// https://groups.google.com/forum/#!msg/golang-nuts/XqW1qcuZgKg/Ui3nQkeLV80J
	if s.command == nil {
		err = commonerrors.UndefinedVariable("command")
		return
	}
	err = s.command.Check()
	if err != nil {
		return
	}
	if s.messaging == nil {
		err = commonerrors.ErrNoLogger
		return
	}
	err = s.messaging.Check()
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

// Wait waits for the command to stop existing on the system.
// This allows check to work if the underlying process was stopped.
func (s *Subprocess) Wait(ctx context.Context) (err error) {
	var pid int
	if s.command != nil && s.command.cmdWrapper.cmd != nil && s.command.cmdWrapper.cmd.Process != nil {
		pid = s.command.cmdWrapper.cmd.Process.Pid
	} else {
		return commonerrors.New(commonerrors.ErrConflict, "command not started")
	}

	return proc.WaitForCompletion(ctx, pid)
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
		s.messaging.LogFailedStart(err)
		s.isRunning.Store(false)
		s.Cancel()
		return
	}
	pid, err := cmd.Pid()
	if err != nil {
		s.messaging.LogFailedStart(err)
		s.isRunning.Store(false)
		s.Cancel()
		return
	}

	s.isRunning.Store(true)
	s.messaging.SetPid(pid)
	s.messaging.LogStarted()
	return
}

// Cancel interrupts an ongoing process. This method is idempotent.
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
		return commonerrors.New(commonerrors.ErrConflict, "process is already started")
	}
	s.processMonitoring.Reset()
	s.command.Reset()
	s.messaging.LogStart()
	s.runProcessMonitoring()
	cmd := s.getCmd()
	s.isRunning.Store(true)
	err = cmd.Run()
	s.isRunning.Store(false)
	s.messaging.LogEnd(err)
	return
}

// Stop stops the process straight away if currently working without waiting for completion. This method should be used in combination with `Start`.
// However, in order to interrupt a process however it was started (using `Start` or `Execute`), prefer `Cancel`.
// This method is idempotent.
func (s *Subprocess) Stop() (err error) {
	return s.stop(true)
}

// Interrupt terminates the process
// This method should be used in combination with `Start`.
// This method is idempotent
func (s *Subprocess) Interrupt(ctx context.Context) (err error) {
	return s.interrupt(ctx)
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
	s.messaging.LogStopping()
	err = s.getCmd().Stop()
	s.command.Reset()
	s.isRunning.Store(false)
	s.messaging.LogEnd(nil)
	return
}

func (s *Subprocess) interrupt(ctx context.Context) (err error) {
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
	s.messaging.LogStopping()
	err = s.getCmd().Interrupt(ctx)
	s.command.Reset()
	s.isRunning.Store(false)
	s.messaging.LogEnd(nil)
	return
}
