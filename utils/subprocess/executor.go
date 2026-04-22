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
	return NewWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// NewWithDir creates a subprocess description with a working directory.
func NewWithDir(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (*Subprocess, error) {
	return NewWithEnvironmentWithDir(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// NewWithEnvironment creates a subprocess description. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func NewWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	return NewWithEnvironmentWithDir(ctx, loggers, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// NewWithEnvironmentWithDir creates a subprocess description with a working directory. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func NewWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (p *Subprocess, err error) {
	return NewWithEnvironmentWithIOWithDir(ctx, loggers, nil, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// NewWithIO creates a subprocess description with overridden stdin/stdout/stderr.
func NewWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (*Subprocess, error) {
	return NewWithIOWithDir(ctx, loggers, io, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// NewWithIOWithDir creates a subprocess description with a working directory and overridden stdin/stdout/stderr.
func NewWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (*Subprocess, error) {
	return NewWithEnvironmentWithIOWithDir(ctx, loggers, io, nil, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// NewWithEnvironmentWithIO creates a subprocess description with overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func NewWithEnvironmentWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	return NewWithEnvironmentWithIOWithDir(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// NewWithEnvironmentWithIOWithDir creates a subprocess description with a working directory and overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func NewWithEnvironmentWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (p *Subprocess, err error) {
	p, err = newSubProcess(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Me(), dir, cmd, args...)
	return
}

func newSubProcess(ctx context.Context, loggers logs.Loggers, io ICommandIO, env []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.SetupAsWithEnvironmentWithCustomIOWithDir(ctx, loggers, io, env, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
	return
}

func newPlainSubProcess(ctx context.Context, loggers logs.Loggers, env []string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (p *Subprocess, err error) {
	p = new(Subprocess)
	err = p.setup(ctx, loggers, nil, env, false, "", "", "", as, dir, cmd, args...)
	return
}

// ExecuteWithEnvironment executes a command (i.e. spawns a subprocess). It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return ExecuteWithEnvironmentWithDir(ctx, loggers, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// ExecuteWithEnvironmentWithDir executes a command (i.e. spawns a subprocess) with a working directory. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (err error) {
	return ExecuteWithEnvironmentWithIOWithDir(ctx, loggers, nil, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// StartWithEnvironment starts a command (i.e. spawns a subprocess). It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func StartWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (*Subprocess, error) {
	return StartWithEnvironmentWithDir(ctx, loggers, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// StartWithEnvironmentWithDir starts a command (i.e. spawns a subprocess) with a working directory. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func StartWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (*Subprocess, error) {
	return StartWithEnvironmentWithIOWithDir(ctx, loggers, nil, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// Execute executes a command (i.e. spawns a subprocess).
func Execute(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) error {
	return ExecuteWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// ExecuteWithDir executes a command (i.e. spawns a subprocess) with a working directory.
func ExecuteWithDir(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) error {
	return ExecuteWithEnvironmentWithDir(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// Start starts a command (i.e. spawns a subprocess).
func Start(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (*Subprocess, error) {
	return StartWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// StartWithDir starts a command (i.e. spawns a subprocess) with a working directory.
func StartWithDir(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (*Subprocess, error) {
	return StartWithEnvironmentWithDir(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// ExecuteAs executes a command (i.e. spawns a subprocess) as a different user.
func ExecuteAs(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) error {
	return ExecuteAsWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// ExecuteAsWithDir executes a command (i.e. spawns a subprocess) as a different user with a working directory.
func ExecuteAsWithDir(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) error {
	return ExecuteAsWithEnvironmentWithDir(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
}

// ExecuteAsWithEnvironment executes a command (i.e. spawns a subprocess) as a different user. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteAsWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return ExecuteAsWithEnvironmentWithDir(ctx, loggers, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// ExecuteAsWithEnvironmentWithDir executes a command (i.e. spawns a subprocess) as a different user with a working directory. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteAsWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (err error) {
	return ExecuteAsWithEnvironmentWithIOWithDir(ctx, loggers, nil, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
}

// ExecuteWithSudo executes a command (i.e. spawns a subprocess) as root.
func ExecuteWithSudo(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) error {
	return ExecuteWithSudoWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// ExecuteWithSudoWithDir executes a command (i.e. spawns a subprocess) as root with a working directory.
func ExecuteWithSudoWithDir(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) error {
	return ExecuteAsWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Sudo(), dir, cmd, args...)
}

// ExecuteWithEnvironmentWithIO executes a command (i.e. spawns a subprocess) with overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteWithEnvironmentWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return ExecuteWithEnvironmentWithIOWithDir(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// ExecuteWithEnvironmentWithIOWithDir executes a command (i.e. spawns a subprocess) with a working directory and overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteWithEnvironmentWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (err error) {
	p, err := NewWithEnvironmentWithIOWithDir(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

// StartWithEnvironmentWithIO starts a command (i.e. spawns a subprocess) with overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func StartWithEnvironmentWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (p *Subprocess, err error) {
	return StartWithEnvironmentWithIOWithDir(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// StartWithEnvironmentWithIOWithDir starts a command (i.e. spawns a subprocess) with a working directory and overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func StartWithEnvironmentWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (p *Subprocess, err error) {
	p, err = NewWithEnvironmentWithIOWithDir(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
	if err != nil {
		return
	}
	err = p.Start()
	return
}

// ExecuteWithIO executes a command (i.e. spawns a subprocess) with overridden stdin/stdout/stderr.
func ExecuteWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) error {
	return ExecuteWithIOWithDir(ctx, loggers, io, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// ExecuteWithIOWithDir executes a command (i.e. spawns a subprocess) with a working directory and overridden stdin/stdout/stderr.
func ExecuteWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) error {
	return ExecuteWithEnvironmentWithIOWithDir(ctx, loggers, io, nil, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// ExecuteAsWithIO executes a command (i.e. spawns a subprocess) as a different user with overridden stdin/stdout/stderr.
func ExecuteAsWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) error {
	return ExecuteAsWithIOWithDir(ctx, loggers, io, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// ExecuteAsWithIOWithDir executes a command (i.e. spawns a subprocess) as a different user with a working directory and overridden stdin/stdout/stderr.
func ExecuteAsWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) error {
	return ExecuteAsWithEnvironmentWithIOWithDir(ctx, loggers, io, nil, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
}

// ExecuteAsWithEnvironmentWithIO executes a command (i.e. spawns a subprocess) as a different user with overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteAsWithEnvironmentWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return ExecuteAsWithEnvironmentWithIOWithDir(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// ExecuteAsWithEnvironmentWithIOWithDir executes a command (i.e. spawns a subprocess) as a different user with a working directory and overridden stdin/stdout/stderr. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func ExecuteAsWithEnvironmentWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnvVars []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (err error) {
	p, err := newSubProcess(ctx, loggers, io, additionalEnvVars, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
	if err != nil {
		return
	}
	return p.Execute()
}

// ExecuteWithSudoWithIO executes a command (i.e. spawns a subprocess) as root with overridden stdin/stdout/stderr.
func ExecuteWithSudoWithIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) error {
	return ExecuteWithSudoWithIOWithDir(ctx, loggers, io, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// ExecuteWithSudoWithIOWithDir executes a command (i.e. spawns a subprocess) as root with a working directory and overridden stdin/stdout/stderr.
func ExecuteWithSudoWithIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) error {
	return ExecuteAsWithIOWithDir(ctx, loggers, io, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Sudo(), dir, cmd, args...)
}

// Output executes a command and returns its output (stdOutput and stdErr are merged) as string.
func Output(ctx context.Context, loggers logs.Loggers, cmd string, args ...string) (string, error) {
	return OutputWithDir(ctx, loggers, "", cmd, args...)
}

// OutputWithDir executes a command with a working directory and returns its output (stdOutput and stdErr are merged) as string.
func OutputWithDir(ctx context.Context, loggers logs.Loggers, dir string, cmd string, args ...string) (string, error) {
	return OutputWithEnvironmentWithDir(ctx, loggers, nil, dir, cmd, args...)
}

// OutputWithEnvironment executes a command and returns its output (stdOutput and stdErr are merged) as string. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func OutputWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, cmd string, args ...string) (string, error) {
	return OutputWithEnvironmentWithDir(ctx, loggers, additionalEnvVars, "", cmd, args...)
}

// OutputWithEnvironmentWithDir executes a command with a working directory and returns its output (stdOutput and stdErr are merged) as string. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func OutputWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, dir string, cmd string, args ...string) (string, error) {
	return OutputAsWithEnvironmentWithDir(ctx, loggers, additionalEnvVars, commandUtils.Me(), dir, cmd, args...)
}

// OutputAs executes a command as a different user and returns its output (stdOutput and stdErr are merged) as string.
func OutputAs(ctx context.Context, loggers logs.Loggers, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (string, error) {
	return OutputAsWithDir(ctx, loggers, as, "", cmd, args...)
}

// OutputAsWithDir executes a command as a different user with a working directory and returns its output (stdOutput and stdErr are merged) as string.
func OutputAsWithDir(ctx context.Context, loggers logs.Loggers, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (string, error) {
	return OutputAsWithEnvironmentWithDir(ctx, loggers, nil, as, dir, cmd, args...)
}

// OutputAsWithEnvironment executes a command as a different user and returns its output (stdOutput and stdErr are merged) as string. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func OutputAsWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (output string, err error) {
	return OutputAsWithEnvironmentWithDir(ctx, loggers, additionalEnvVars, as, "", cmd, args...)
}

// OutputAsWithEnvironmentWithDir executes a command as a different user with a working directory and returns its output (stdOutput and stdErr are merged) as string. It allows to specify the environment the subprocess should use. Each entry is of the form "key=value".
func OutputAsWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnvVars []string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (output string, err error) {
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
	p, err := newPlainSubProcess(ctx, mLoggers, additionalEnvVars, as, dir, cmd, args...)
	if err != nil {
		return
	}
	err = p.Execute()
	output = stringLogger.GetLogContent()
	return
}

// Setup sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure.
func (s *Subprocess) Setup(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.SetupWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// SetupWithDir sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure with a working directory.
func (s *Subprocess) SetupWithDir(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (err error) {
	return s.SetupWithEnvironmentWithDir(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// SetupWithEnvironment sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure. Compared to Setup, it allows specifying additional environment variables to be used by the process.
func (s *Subprocess) SetupWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.SetupWithEnvironmentWithDir(ctx, loggers, additionalEnv, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// SetupWithEnvironmentWithDir sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure with a working directory. Compared to SetupWithDir, it allows specifying additional environment variables to be used by the process.
func (s *Subprocess) SetupWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, nil, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Me(), dir, cmd, args...)
}

// SetupAs is similar to Setup but allows the command to be run as a different user.
func (s *Subprocess) SetupAs(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.SetupAsWithDir(ctx, loggers, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// SetupAsWithDir is similar to SetupWithDir but allows the command to be run as a different user.
func (s *Subprocess) SetupAsWithDir(ctx context.Context, loggers logs.Loggers, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (err error) {
	return s.SetupAsWithEnvironmentWithDir(ctx, loggers, nil, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
}

// SetupAsWithEnvironment is similar to Setup but allows the command to be run as a different user. Compared to SetupAs, it allows specifying additional environment variables to be used by the process.
func (s *Subprocess) SetupAsWithEnvironment(ctx context.Context, loggers logs.Loggers, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.SetupAsWithEnvironmentWithDir(ctx, loggers, additionalEnv, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// SetupAsWithEnvironmentWithDir is similar to SetupWithDir but allows the command to be run as a different user. Compared to SetupAsWithDir, it allows specifying additional environment variables to be used by the process.
func (s *Subprocess) SetupAsWithEnvironmentWithDir(ctx context.Context, loggers logs.Loggers, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, nil, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
}

// SetupWithCustomIO sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.SetupWithCustomIOWithDir(ctx, loggers, io, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// SetupWithCustomIOWithDir sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure with a working directory. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupWithCustomIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (err error) {
	return s.SetupWithEnvironmentWithCustomIOWithDir(ctx, loggers, io, nil, messageOnStart, messageOnSuccess, messageOnFailure, dir, cmd, args...)
}

// SetupWithEnvironmentWithCustomIO sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure. Compared to SetupWithCustomIO, it allows specifying additional environment variables to be used by the process. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupWithEnvironmentWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, cmd string, args ...string) (err error) {
	return s.SetupWithEnvironmentWithCustomIOWithDir(ctx, loggers, io, additionalEnv, messageOnStart, messageOnSuccess, messageOnFailure, "", cmd, args...)
}

// SetupWithEnvironmentWithCustomIOWithDir sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure with a working directory. Compared to SetupWithCustomIOWithDir, it allows specifying additional environment variables to be used by the process. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupWithEnvironmentWithCustomIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, dir string, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, io, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, commandUtils.Me(), dir, cmd, args...)
}

// SetupAsWithCustomIO is similar to SetupWithCustomIO but allows the command to be run as a different user. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupAsWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.SetupAsWithCustomIOWithDir(ctx, loggers, io, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// SetupAsWithCustomIOWithDir is similar to SetupWithCustomIOWithDir but allows the command to be run as a different user. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupAsWithCustomIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (err error) {
	return s.SetupAsWithEnvironmentWithCustomIOWithDir(ctx, loggers, io, nil, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
}

// SetupAsWithEnvironmentWithCustomIO is similar to SetupWithCustomIO but allows the command to be run as a different user. Compared to SetupAsWithCustomIO, it allows specifying additional environment variables to be used by the process. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupAsWithEnvironmentWithCustomIO(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, cmd string, args ...string) (err error) {
	return s.SetupAsWithEnvironmentWithCustomIOWithDir(ctx, loggers, io, additionalEnv, messageOnStart, messageOnSuccess, messageOnFailure, as, "", cmd, args...)
}

// SetupAsWithEnvironmentWithCustomIOWithDir is similar to SetupWithCustomIOWithDir but allows the command to be run as a different user. Compared to SetupAsWithCustomIOWithDir, it allows specifying additional environment variables to be used by the process. It allows the stdin, stdout, and stderr to be overridden.
func (s *Subprocess) SetupAsWithEnvironmentWithCustomIOWithDir(ctx context.Context, loggers logs.Loggers, io ICommandIO, additionalEnv []string, messageOnStart string, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (err error) {
	return s.setup(ctx, loggers, io, additionalEnv, true, messageOnStart, messageOnSuccess, messageOnFailure, as, dir, cmd, args...)
}

// Setup sets up a sub-process i.e. defines the command cmd and the messages on start, success and failure as well as the stdin, stdout, and stderr.
func (s *Subprocess) setup(ctx context.Context, loggers logs.Loggers, io ICommandIO, env []string, withAdditionalMessages bool, messageOnStart, messageOnSuccess, messageOnFailure string, as *commandUtils.CommandAsDifferentUser, dir string, cmd string, args ...string) (err error) {
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
		s.command = newCommandWithCustomIO(loggers, io, as, env, dir, cmd, args...)
	} else {
		s.command = newCommand(loggers, as, env, dir, cmd, args...)
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
