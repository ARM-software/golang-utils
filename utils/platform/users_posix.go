//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm darwin aix dragonfly freebsd nacl netbsd openbsd solaris

package platform

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

var (
	// SudoCommand describes the command to use to execute command as root
	// when running in Docker, change to [gosu](https://github.com/tianon/gosu)
	SudoCommand = "sudo"
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
	cmd := exec.CommandContext(ctx, SudoCommand, "useradd", "-U", "-m", username, "-p", pwd, "-c", comment)
	err = runCommand("useradd", cmd)
	if err != nil {
		return
	}
	cmd = exec.CommandContext(ctx, SudoCommand, "passwd", "-d", username)
	err = runCommand("passwd", cmd)
	return

}

func removeUser(ctx context.Context, username string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "userdel", "-f", "-r", username)
	err = runCommand("userdel", cmd)
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "groupadd", "-f", groupName)
	err = runCommand("groupadd", cmd)
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "groupdel", groupName)
	err = runCommand("groupdel", cmd)
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "usermod", "-a", "-G", groupName, username)
	err = runCommand("usermod", cmd)
	return
}

func dissociateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "gpasswd", "-d", username, groupName)
	err = runCommand("gpasswd", cmd)
	return
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
