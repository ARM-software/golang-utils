package proc

import (
	"context"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

//go:generate go run github.com/dmarkham/enumer -type=InterruptType -text -json -yaml
type InterruptType int

const (
	SigInt                           InterruptType = 2
	SigKill                          InterruptType = 9
	SigTerm                          InterruptType = 15
	SubprocessTerminationGracePeriod               = 10 * time.Millisecond
)

func InterruptProcess(ctx context.Context, pid int, signal InterruptType) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	process, err := FindProcess(ctx, pid)
	if err != nil || process == nil {
		err = commonerrors.Ignore(err, commonerrors.ErrNotFound)
		return
	}

	switch signal {
	case SigInt:
		err = process.Interrupt(ctx)
	case SigKill:
		err = process.KillWithChildren(ctx)
	case SigTerm:
		err = process.Terminate(ctx)
	default:
		err = commonerrors.New(commonerrors.ErrInvalid, "unknown interrupt type for process")
	}

	err = commonerrors.IgnoreCorrespondTo(err, "process already finished") // equivalent to desired state
	return
}

// TerminateGracefullyWithChildren follows the pattern set by [kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination) and terminates processes gracefully by first sending a SIGINT+SIGTERM and then a SIGKILL after the grace period has elapsed.
// It does not attempt to terminate the process group. If you wish to terminate the process group directly then send -pgid to TerminateGracefully but
// this does not guarantee that the group will be terminated gracefully.
// Instead, this function lists each child and attempts to kill them gracefully concurrently. It will then attempt to gracefully terminate itself.
// Due to the multi-stage process and the fact that the full grace period must pass for each stage specified above, the total maximum length of this
// function will be 2*gracePeriod not gracePeriod.
func TerminateGracefullyWithChildren(ctx context.Context, pid int, gracePeriod time.Duration) (err error) {
	return terminateGracefullyWithChildren(ctx, pid, gracePeriod, TerminateGracefully)
}

// TerminateGracefullyWithChildrenWithSpecificInterrupts follows the pattern set by [kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination) and terminates processes gracefully by first sending the specified interrupts and then a SIGKILL after the grace period has elapsed.
// It does not attempt to terminate the process group. If you wish to terminate the process group directly then send -pgid to TerminateGracefully but
// this does not guarantee that the group will be terminated gracefully.
// Instead, this function lists each child and attempts to kill them gracefully concurrently. It will then attempt to gracefully terminate itself.
// Due to the multi-stage process and the fact that the full grace period must pass for each stage specified above, the total maximum length of this
// function will be 2*gracePeriod not gracePeriod.
func TerminateGracefullyWithChildrenWithSpecificInterrupts(ctx context.Context, pid int, gracePeriod time.Duration, interrupts ...InterruptType) (err error) {
	return terminateGracefullyWithChildren(ctx, pid, gracePeriod, func(ctx context.Context, pid int, gracePeriod time.Duration) error {
		return TerminateGracefullyWithSpecificInterrupts(ctx, pid, gracePeriod, interrupts...)
	})
}

func terminateGracefullyWithChildren(ctx context.Context, pid int, gracePeriod time.Duration, terminate func(context.Context, int, time.Duration) error) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	p, err := FindProcess(ctx, pid)
	if err != nil {
		if commonerrors.Any(err, commonerrors.ErrNotFound) {
			err = nil
			return
		}

		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "an error occurred whilst searching for process '%v'", pid)
		return
	}

	children, err := p.Children(ctx)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not check for children for pid '%v'", pid)
		return
	}

	if len(children) == 0 {
		err = terminate(ctx, pid, gracePeriod)
		return
	}

	childGroup, terminateCtx := errgroup.WithContext(ctx)
	childGroup.SetLimit(len(children))
	for _, child := range children {
		if child.IsRunning() {
			childGroup.Go(func() error {
				return terminateGracefullyWithChildren(terminateCtx, child.Pid(), gracePeriod, terminate)
			})
		}
	}
	err = childGroup.Wait()
	if err != nil {
		return
	}

	err = terminate(ctx, pid, gracePeriod)
	return
}

func terminateGracefully(ctx context.Context, pid int, gracePeriod time.Duration, interrupts ...InterruptType) (err error) {
	if len(interrupts) == 0 {
		err = commonerrors.New(commonerrors.ErrInvalid, "at least one interrupt must be provided")
		return
	}

	for _, interrupt := range interrupts {
		err = InterruptProcess(ctx, pid, interrupt)
		if err != nil {
			return
		}
	}

	return parallelisation.RunActionWithParallelCheck(ctx,
		func(ctx context.Context) error {
			parallelisation.SleepWithContext(ctx, gracePeriod)
			return nil
		},
		func(ctx context.Context) bool {
			_, fErr := FindProcess(ctx, pid)
			return commonerrors.Any(fErr, commonerrors.ErrNotFound)

		}, 200*time.Millisecond)
}

// TerminateGracefully follows the pattern set by [kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination) and terminates processes gracefully by first sending a SIGINT+SIGTERM and then a SIGKILL after the grace period has elapsed.
func TerminateGracefully(ctx context.Context, pid int, gracePeriod time.Duration) (err error) {
	defer func() { _ = InterruptProcess(context.Background(), pid, SigKill) }()
	_ = terminateGracefully(ctx, pid, gracePeriod, SigInt, SigTerm)
	err = InterruptProcess(ctx, pid, SigKill)
	return
}

// TerminateGracefullyWithSpecificInterrupts follows the pattern set by [kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination) and terminates processes gracefully by first sending the specified interrupts and then a SIGKILL after the grace period has elapsed.
func TerminateGracefullyWithSpecificInterrupts(ctx context.Context, pid int, gracePeriod time.Duration, interrupts ...InterruptType) (err error) {
	defer func() { _ = InterruptProcess(context.Background(), pid, SigKill) }()
	_ = terminateGracefully(ctx, pid, gracePeriod, interrupts...)
	err = InterruptProcess(ctx, pid, SigKill)
	return
}

// CancelExecCommand defines a more robust way to cancel subprocesses than what is done per default by [CommandContext](https://pkg.go.dev/os/exec#CommandContext)
func CancelExecCommand(cmd *exec.Cmd) (err error) {
	if cmd == nil {
		err = commonerrors.UndefinedVariable("command")
		return
	}
	if cmd.Process == nil {
		return
	}
	err = TerminateGracefully(context.Background(), cmd.Process.Pid, SubprocessTerminationGracePeriod)
	err = commonerrors.Ignore(err, os.ErrProcessDone)
	if err != nil {
		// Default behaviour
		err = cmd.Process.Kill()
	}
	return
}

// DefineCmdCancel sets and overwrites the cmd.Cancel function with CancelExecCommand so that it is more robust and thorough.
func DefineCmdCancel(cmd *exec.Cmd) (*exec.Cmd, error) {
	if cmd == nil {
		return nil, commonerrors.UndefinedVariable("command")
	}
	cmd.Cancel = func() error {
		return CancelExecCommand(cmd)
	}
	return cmd, nil
}

// WaitForCompletion will wait for a given process to complete.
// This allows check to work if the underlying process was stopped without needing the os.Process that started it.
func WaitForCompletion(ctx context.Context, pid int) (err error) {
	pids, err := getGroupProcesses(ctx, pid)
	err = commonerrors.Ignore(err, commonerrors.ErrNotFound)
	if err != nil {
		return
	}
	return parallelisation.WaitUntil(ctx, func(ctx2 context.Context) (allStopped bool, err error) {
		allStopped = true
		err = collection.ForAll(pids, func(subPid int) (subErr error) {
			proc, subErr := FindProcess(ctx2, subPid)
			switch {
			case subErr == nil:
				if proc.IsRunning() && !proc.IsAZombie() {
					allStopped = false
				}
			case commonerrors.Any(subErr, commonerrors.ErrNotFound):
				// gone so stopped
			default:
				subErr = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, subErr, "an error occurred whilst finding sub process '%v'", subPid)
				return
			}
			return
		})
		return
	}, 200*time.Millisecond)
}
