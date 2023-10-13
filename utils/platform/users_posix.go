//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm darwin aix dragonfly freebsd nacl netbsd openbsd solaris

package platform

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func addUser(ctx context.Context, username, fullname, password string) (err error) {
	pwd := password
	if pwd == "" {
		pwd = "tmp123"
	}
	comment := fullname
	if comment == "" {
		comment = username
	}
	err = executeCommand(ctx, "useradd", "-U", "-m", username, "-p", pwd, "-c", comment)
	if err != nil {
		return
	}
	err = executeCommand(ctx, "passwd", "-d", username)
	return

}

func removeUser(ctx context.Context, username string) (err error) {
	err = executeCommand(ctx, "userdel", "-f", "-r", username)
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	err = executeCommand(ctx, "groupadd", "-f", groupName)
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	err = executeCommand(ctx, "groupdel", groupName)
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	err = executeCommand(ctx, "usermod", "-a", "-G", groupName, username)
	return
}

func dissociateUserFromGroup(ctx context.Context, username, groupName string) (err error) {
	err = executeCommand(ctx, "gpasswd", "-d", username, groupName)
	return
}

func executeCommand(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: missing command to execute", commonerrors.ErrUndefined)
	}
	cmdName, cmdArgs := defineCommandWithPrivileges(args...)
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	return runCommand(args[0], cmd)
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
