//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm darwin aix dragonfly freebsd nacl netbsd openbsd solaris

package platform

import "github.com/ARM-software/golang-utils/utils/subprocess/command"

var (
	// sudoCommand describes the command to use to execute command as root
	// when running in Docker, change to [gosu root](https://github.com/tianon/gosu)
	sudoCommand = command.Sudo()
)

// DefineSudoCommand defines the command to run to be `root` or a user with enough privileges to manage accounts.
// e.g.
//   - args="sudo" to run commands as `root`
//   - args="su", "tom" if `tom` has enough privileges to run the command
//   - args="gosu", "tom" if `tom` has enough privileges to run the command in a container and `gosu` is installed
func DefineSudoCommand(args ...string) {
	sudoCommand = command.NewCommandAsDifferentUser(args...)
}

func defineCommandWithPrivileges(args ...string) (string, []string) {
	return sudoCommand.RedefineCommand(args...)
}
