//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm darwin aix dragonfly freebsd nacl netbsd openbsd solaris

package platform

import (
	"os/exec"

	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

var (
	// sudoCommand describes the command to use to execute command as root
	// when running in Docker, change to [gosu root](https://github.com/tianon/gosu)
	sudoCommand = command.Sudo()
)

// DefineSudoCommand defines the command to run to be `root` or a user with enough privileges (superuser).
// e.g.
//   - args="sudo" to run commands as `root`
//   - args="su", "tom" if `tom` has enough privileges to run the command
func DefineSudoCommand(args ...string) {
	sudoCommand = command.NewCommandAsDifferentUser(args...)
}

func getRunCommandWithPrivileges() *command.CommandAsDifferentUser {
	return sudoCommand
}

func hasPasswordlessPrivileges() bool {
	// See https://www.baeldung.com/linux/sudo-passwordless-check
	return exec.Command("sh", "-c", "sudo -n true 2>/dev/null || exit 1").Run() == nil
}
