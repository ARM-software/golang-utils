package command

import "strings"

// CommandAsDifferentUser helps redefining commands so that they are run as a different user or with more privileges.
type CommandAsDifferentUser struct {
	// changeUserCmd describes the command to use to execute any command as a different user
	// e.g. it can be set as "sudo" to run commands as `root` or as "su","tom" or "gosu","jack"
	changeUserCmd []string
}

// Redefine redefines a command so that it will be run as a different user.
func (c *CommandAsDifferentUser) Redefine(cmd string, args ...string) (cmdName string, cmdArgs []string) {
	newArgs := []string{cmd}
	newArgs = append(newArgs, args...)
	cmdName, cmdArgs = c.RedefineCommand(newArgs...)
	return
}

// RedefineCommand is the same as Redefine but with no separation between the command and its arguments (like the command in Docker)
func (c *CommandAsDifferentUser) RedefineCommand(args ...string) (cmdName string, cmdArgs []string) {
	if len(c.changeUserCmd) > 0 {
		cmdName = c.changeUserCmd[0]
		for i := 1; i < len(c.changeUserCmd); i++ {
			cmdArgs = append(cmdArgs, c.changeUserCmd[i])
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

// RedefineInShellForm returns the new command defined in shell form.
func (c *CommandAsDifferentUser) RedefineInShellForm(cmd string, args ...string) string {
	ncmd, nargs := c.Redefine(cmd, args...)
	return AsShellForm(ncmd, nargs...)
}

// NewCommandAsDifferentUser defines a command wrapper which helps redefining commands so that they are run as a different user.
// e.g.
//   - switchUserCmd="sudo" to run commands as `root`
//   - switchUserCmd="su", "tom" if `tom` has enough privileges to run the command
//   - switchUserCmd="gosu", "tom" if `tom` has enough privileges to run the command in a container and `gosu` is installed
func NewCommandAsDifferentUser(switchUserCmd ...string) *CommandAsDifferentUser {
	return &CommandAsDifferentUser{changeUserCmd: switchUserCmd}
}

// NewCommandAsRoot will create a command translator which will run command with `sudo`
func NewCommandAsRoot() *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("sudo")
}

// Sudo will call commands with `sudo`. Similar to NewCommandAsRoot
func Sudo() *CommandAsDifferentUser {
	return NewCommandAsRoot()
}

// NewCommandInContainerAs will redefine commands to be run in containers as `username`. It will expect [gosu](https://github.com/tianon/gosu) to be installed and the user to have been defined.
func NewCommandInContainerAs(username string) *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("gosu", username)
}

// Gosu is similar to NewCommandInContainerAs.
func Gosu(username string) *CommandAsDifferentUser {
	return NewCommandInContainerAs(username)
}

// Su will run commands as the user username using [su](https://www.unix.com/man-page/posix/1/su/)
func Su(username string) *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("su", username)
}

// Me will run the commands without switching user. It is a no operation wrapper.
func Me() *CommandAsDifferentUser {
	return NewCommandAsDifferentUser()
}

// AsShellForm returns a command in its shell form.
func AsShellForm(cmd string, args ...string) string {
	newCmd := []string{cmd}
	newCmd = append(newCmd, args...)
	return strings.Join(newCmd, " ")
}
