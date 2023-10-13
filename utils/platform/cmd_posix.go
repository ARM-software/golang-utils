//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm darwin aix dragonfly freebsd nacl netbsd openbsd solaris

package platform

var (
	// sudoCommand describes the command to use to execute command as root
	// when running in Docker, change to [gosu root](https://github.com/tianon/gosu)
	sudoCommand = []string{"sudo"}
)

// DefineSudoCommand defines the command to run to be `root` or a user with enough privileges to manage accounts.
// e.g.
//   - args="sudo" to run commands as `root`
//   - args="su", "tom" if `tom` has enough privileges to run the command
//   - args="gosu", "tom" if `tom` has enough privileges to run the command in a container and `gosu` is installed
func DefineSudoCommand(args ...string) {
	sudoCommand = args
}

// DefineCommandWithPrivileges redefines a command so that it is run with privileges or as `root` depending how `DefineSudoCommand` was called.
func DefineCommandWithPrivileges(cmd string, args ...string) (cmdName string, cmdArgs []string) {
	newArgs := []string{cmd}
	newArgs = append(newArgs, args...)
	cmdName, cmdArgs = defineCommandWithPrivileges(newArgs...)
	return
}

func defineCommandWithPrivileges(args ...string) (cmdName string, cmdArgs []string) {
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
	return
}
