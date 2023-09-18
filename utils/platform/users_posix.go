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
	err = convertCommandError("useradd", cmd.Run())
	if err != nil {
		return
	}
	cmd = exec.CommandContext(ctx, SudoCommand, "passwd", "-d", username)
	err = convertCommandError("passwd", cmd.Run())
	return

}

func removeUser(ctx context.Context, username string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "userdel", "-f", "-r", "-Z", username)
	err = convertCommandError("userdel", cmd.Run())
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "groupadd", "-f", groupName)
	err = convertCommandError("groupadd", cmd.Run())
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "groupdel", groupName)
	err = convertCommandError("groupdel", cmd.Run())
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "usermod ", "-a", "-G", groupName, username)
	err = convertCommandError("usermod", cmd.Run())
	return
}

func dissociateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "gpasswd", "-d", username, groupName)
	err = convertCommandError("gpasswd", cmd.Run())
	return
}

func convertCommandError(cmd string, err error) error {
	if err == nil {
		return nil
	}
	newErr := commonerrors.ConvertContextError(err)
	switch {
	case commonerrors.Any(newErr, commonerrors.ErrTimeout, commonerrors.ErrCancelled, commonerrors.ErrUnknown, commonerrors.ErrUnsupported, commonerrors.ErrCondition, commonerrors.ErrForbidden):
		return newErr
	default:
		newErr = fmt.Errorf("%w: the command `%v` failed: %v", commonerrors.ErrUnknown, cmd, err.Error())
		return newErr
	}
}
