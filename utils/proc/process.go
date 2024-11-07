/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package proc

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

const (
	statusRunning = "running"
	statusSleep   = "sleep"
	statusIdle    = "idle"
)

// Ps returns all processes in a similar fashion to `ps` command on Unix.
//
// This of course will be a point-in-time snapshot of when this method was
// called. Some operating systems don't provide snapshot capability of the
// process table, in which case the process table returned might contain
// ephemeral entities that happened to be running when this was called.
func Ps(ctx context.Context) (processes []IProcess, err error) {
	pss, err := process.ProcessesWithContext(ctx)
	err = ConvertProcessError(err)
	if err != nil {
		return
	}
	for i := range pss {
		processes = append(processes, wrapProcess(pss[i]))
	}
	return
}

// FindProcess looks up a single process by pid.
//
// Process will be nil and error will be commonerrors.ErrNotFound if a matching process is
// not found.
func FindProcess(ctx context.Context, pid int) (p IProcess, err error) {
	p, err = NewProcess(ctx, pid)
	if commonerrors.Any(err, nil, commonerrors.ErrTimeout, commonerrors.ErrCancelled) {
		return
	}
	err = fmt.Errorf("%w: process (#%v) could not be found: %v", commonerrors.ErrNotFound, pid, err.Error())
	return
}

// IsProcessRunning states whether a process is running or not. An error is returned if the context is Done while looking for the process state.
func IsProcessRunning(ctx context.Context, pid int) (running bool, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	p, SubErr := FindProcess(ctx, pid)
	switch {
	case SubErr == nil:
		running = p.IsRunning()
	case commonerrors.Any(SubErr, commonerrors.ErrTimeout, commonerrors.ErrCancelled):
		err = SubErr
	default:
		running = false
	}
	return
}

type ps struct {
	imp *process.Process
}

func (p *ps) IsRunning() (running bool) {
	running = isProcessRunning(p.imp)
	return
}

func (p *ps) Cmdline() string {
	cmd, _ := p.imp.Cmdline()
	return cmd
}

func (p *ps) Cwd() string {
	cwd, _ := p.imp.Cwd()
	return cwd
}

func (p *ps) Parent() IProcess {
	pp, _ := p.imp.Parent()
	return wrapProcess(pp)
}

func (p *ps) Children(ctx context.Context) (children []IProcess, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	cp, err := p.imp.Children()
	err = ConvertProcessError(err)
	if err != nil {
		return
	}
	for i := range cp {
		children = append(children, wrapProcess(cp[i]))
	}
	return
}

func (p *ps) Name() string {
	n, _ := p.imp.Name()
	return n
}

func (p *ps) Environ(ctx context.Context) (env []string, err error) {
	env, err = p.imp.EnvironWithContext(ctx)
	err = ConvertProcessError(err)
	return
}

func (p *ps) Pid() int {
	return int(p.imp.Pid)
}

func (p *ps) PPid() int {
	ppid, _ := p.imp.Ppid()
	return int(ppid)
}

func (p *ps) Executable() string {
	ex, _ := p.imp.Exe()
	return ex
}

func (p *ps) Terminate(ctx context.Context) error {
	err := ConvertProcessError(p.imp.TerminateWithContext(ctx))
	err = commonerrors.Ignore(err, commonerrors.ErrNotFound)
	return err
}

func (p *ps) KillWithChildren(ctx context.Context) error {
	return killProcessAndChildren(ctx, p.imp)
}

func killProcessAndChildren(ctx context.Context, p *process.Process) (err error) {
	// See https://varunksaini.com/posts/kiling-processes-in-go/
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if p == nil {
		return
	}
	defer func() {
		_ = p.Kill()
	}()
	err = ConvertProcessError(p.TerminateWithContext(ctx))
	if err != nil {
		err = commonerrors.Ignore(err, commonerrors.ErrNotFound)
		return
	}
	// First of all, we try to kill the group as it is the preferred/quicker option but requires the processes to have been defined as part of the group
	subErr := ConvertProcessError(killGroup(ctx, p.Pid))
	if subErr != nil {
		subErr = commonerrors.Ignore(subErr, commonerrors.ErrNotFound)
		if subErr == nil {
			err = nil
			return
		}
	}
	err = killChildren(ctx, p)
	if err != nil {
		err = commonerrors.Ignore(err, commonerrors.ErrNotFound)
		return
	}
	if !isProcessRunning(p) {
		return
	}
	err = commonerrors.Ignore(ConvertProcessError(p.KillWithContext(ctx)), commonerrors.ErrNotFound)
	return
}

func killChildren(ctx context.Context, p *process.Process) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if p == nil {
		return
	}
	children, childErr := p.ChildrenWithContext(ctx)
	if childErr != nil {
		childErr = ConvertProcessError(childErr)
		if commonerrors.Any(childErr, commonerrors.ErrTimeout, commonerrors.ErrCancelled) {
			err = childErr
		}
		return
	}
	for i := range children {
		subErr := killProcessAndChildren(ctx, children[i])
		if subErr != nil {
			err = subErr
		}
	}
	if err != nil {
		subErr := killGroup(ctx, p.Pid) // Radical approach of killing the whole process group
		if subErr == nil {
			err = nil
		}
	}
	return err

}

func isProcessRunning(p *process.Process) (running bool) {
	if p == nil {
		return
	}
	running, _ = p.IsRunning()
	if !running {
		return
	}
	// from man 2 kill: If sig is 0, then no signal is sent, but error checking is still performed; this can be used to check for the existence of a process ID or process group ID.
	exist, _ := process.PidExists(p.Pid)
	if !exist {
		running = false
		return
	}
	// On some platforms, such as *nix, a zombie process is reported as a running process by p.IsRunning() but this is not the case. Therefore, a further check is performed on the process status to verify a running process is actually in the expected running state. Nonetheless, status is not cross platform and is not implemented on Windows. For those platform, the status returned by IsRunning is then considered
	status, err := p.Status()
	if err != nil {
		return
	}
	// https://github.com/shirou/gopsutil/blob/e230f528f075f78e713f167c28b692cc15307d19/process/process.go#L48
	_, running = collection.FindInSlice(false, status, statusRunning, statusSleep, statusIdle)
	return
}

// NewProcess creates a new Process instance, it only stores the pid and
// checks that the process exists. Other method on Process can be used
// to get more information about the process. An error will be returned
// if the process does not exist.
func NewProcess(ctx context.Context, pid int) (pr IProcess, err error) {
	p, err := process.NewProcessWithContext(ctx, safecast.ToInt32(pid))
	err = ConvertProcessError(err)
	if err != nil {
		return
	}
	pr = wrapProcess(p)
	return
}

func wrapProcess(p *process.Process) IProcess {
	if p == nil {
		return nil
	}
	return &ps{imp: p}
}
