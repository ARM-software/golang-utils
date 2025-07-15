package proc

import (
	"context"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

//go:generate go run github.com/dmarkham/enumer -type=InterruptType -text -json -yaml
type InterruptType int

const (
	SigInt  InterruptType = 2
	SigKill InterruptType = 9
	SigTerm InterruptType = 15
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
	return
}

// TerminateGracefully follows the pattern set by [kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination) and terminates processes gracefully by first sending a SIGTERM and then a SIGKILL after the grace period has elapsed.
func TerminateGracefully(ctx context.Context, pid int, gracePeriod time.Duration) (err error) {
	defer func() { _ = InterruptProcess(context.Background(), pid, SigKill) }()
	err = InterruptProcess(ctx, pid, SigInt)
	if err != nil {
		return
	}
	err = InterruptProcess(ctx, pid, SigTerm)
	if err != nil {
		return
	}
	_, fErr := FindProcess(ctx, pid)
	if commonerrors.Any(fErr, commonerrors.ErrNotFound) {
		// The process no longer exist.
		// No need to wait the grace period
		return
	}
	parallelisation.SleepWithContext(ctx, gracePeriod)
	err = InterruptProcess(ctx, pid, SigKill)
	return
}
