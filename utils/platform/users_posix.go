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
	// sudoCommand describes the command to use to execute command as root
	// when running in Docker, change to [gosu root](https://github.com/tianon/gosu)
	sudoCommand = []string{"sudo"}
)

// DefineSudoCommand defines the command to run to be `root` or a user with enough privileges to manage accounts.
func DefineSudoCommand(args ...string) {
	sudoCommand = args
}

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

func dissociateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	err = executeCommand(ctx, "gpasswd", "-d", username, groupName)
	return
}

func executeCommand(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: missing command to execute", commonerrors.ErrUndefined)
	}
	cmd := defineCommand(ctx, args...)
	return runCommand(args[0], cmd)
}

func defineCommand(ctx context.Context, args ...string) *exec.Cmd {
	var cmdName string
	var cmdArgs []string
	if len(sudoCommand) > 0 {
		cmdName = sudoCommand[0]
		for i := 1; i < len(sudoCommand); i++ {
			cmdArgs = append(cmdArgs, sudoCommand[i])
		}
		cmdArgs = append(cmdArgs, args...)
	} else {
		cmdName = args[0]
		for i := 1; i < len(args); i++ {
			cmdArgs = append(cmdArgs, args[i])
		}
	}
	return exec.CommandContext(ctx, cmdName, cmdArgs...)
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
