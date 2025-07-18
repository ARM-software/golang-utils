package platform

import (
	"context"
	"errors"
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
		err = commonerrors.WrapErrorf(commonerrors.ErrNotFound, err, "could not find path [%v]", path)
		return
	}
	if fi.IsDir() {
		err = removeDirAs(ctx, WithPrivileges(command.Me()), path)
	} else {
		err = removeFileAs(ctx, WithPrivileges(command.Me()), path)
	}
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not remove the path [%v]", path)
	}
	return
}

func executeCommandAs(ctx context.Context, as *command.CommandAsDifferentUser, args ...string) error {
	if as == nil {
		return commonerrors.UndefinedVariable("command wrapper")
	}
	if len(args) == 0 {
		return commonerrors.UndefinedVariable("command to execute")
	}
	cmdName, cmdArgs := as.RedefineCommand(args...)
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	// setting the following to avoid having hanging subprocesses as described in https://github.com/golang/go/issues/24050
	cmd.WaitDelay = 5 * time.Second
	cmd, err := proc.DefineCmdCancel(cmd)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not set the command cancel function")
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
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			details = string(exitError.Stderr)
		}
		newErr = commonerrors.WrapErrorf(commonerrors.ErrUnknown, err, "the command `%v` failed (%v)", cmdDescription, details)
		return newErr
	}
}
