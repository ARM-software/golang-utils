package platform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/proc"
	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

// RemoveWithPrivileges removes a directory even if it is not owned by user (equivalent to sudo rm -rf). It expects the current user to be a superuser.
func RemoveWithPrivileges(ctx context.Context, path string) (err error) {
	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("%w: could not find path [%v]: %v", commonerrors.ErrNotFound, path, err.Error())
		return
	}
	if fi.IsDir() {
		err = removeDirAs(ctx, WithPrivileges(command.Me()), path)
	} else {
		err = removeFileAs(ctx, WithPrivileges(command.Me()), path)
	}
	if err != nil {
		err = fmt.Errorf("%w: could not remove the path [%v]: %v", commonerrors.ErrUnexpected, path, err.Error())
	}
	return
}

func executeCommandAs(ctx context.Context, as *command.CommandAsDifferentUser, args ...string) error {
	if as == nil {
		return fmt.Errorf("%w: missing command wrapper", commonerrors.ErrUndefined)
	}
	if len(args) == 0 {
		return fmt.Errorf("%w: missing command to execute", commonerrors.ErrUndefined)
	}
	cmdName, cmdArgs := as.RedefineCommand(args...)
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	// setting the following to avoid having hanging subprocesses as described in https://github.com/golang/go/issues/24050
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		p, err := proc.FindProcess(context.Background(), cmd.Process.Pid)
		if err == nil {
			return p.KillWithChildren(context.Background())
		} else {
			// Default behaviour
			return cmd.Process.Kill()
		}
	}
	return runCommand(args[0], cmd)
}

func executeCommand(ctx context.Context, args ...string) error {
	return executeCommandAs(ctx, WithPrivileges(command.Me()), args...)
}

func runCommand(cmdDescription string, cmd *exec.Cmd) error {
	_, err := cmd.Output()
	if err == nil {
		return nil
	}
	newErr := commonerrors.ConvertContextError(err)
	switch {
	case commonerrors.Any(newErr, commonerrors.ErrTimeout, commonerrors.ErrCancelled, commonerrors.ErrUnknown, commonerrors.ErrUnsupported, commonerrors.ErrCondition, commonerrors.ErrForbidden):
		return newErr
	default:
		details := "no further details"
		if exitError, ok := err.(*exec.ExitError); ok {
			details = string(exitError.Stderr)
		}
		newErr = fmt.Errorf("%w: the command `%v` failed: %v (%v)", commonerrors.ErrUnknown, cmdDescription, err.Error(), details)
		return newErr
	}
}
