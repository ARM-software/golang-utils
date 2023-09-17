//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm darwin aix dragonfly freebsd nacl netbsd openbsd solaris

package platform

import (
	"context"
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
	err = commonerrors.ConvertContextError(cmd.Run())
	if err != nil {
		return
	}
	cmd = exec.CommandContext(ctx, SudoCommand, "passwd", "-d", username)
	err = commonerrors.ConvertContextError(cmd.Run())
	return

}

func removeUser(ctx context.Context, username string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "userdel", "-f", "-r", "-Z", username)
	err = commonerrors.ConvertContextError(cmd.Run())
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "groupadd", "-f", groupName)
	err = commonerrors.ConvertContextError(cmd.Run())
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "groupdel", groupName)
	err = commonerrors.ConvertContextError(cmd.Run())
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "usermod ", "-a", "-G", groupName, username)
	err = commonerrors.ConvertContextError(cmd.Run())
	return
}

func dissociateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	cmd := exec.CommandContext(ctx, SudoCommand, "gpasswd", "-d", username, groupName)
	err = commonerrors.ConvertContextError(cmd.Run())
	return
}
