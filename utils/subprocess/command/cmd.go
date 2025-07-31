package command

import (
	"fmt"
	"strings"
	"sync"
)

// CommandAsDifferentUser helps to redefine commands so that they are run as a different user or with more privileges.
type CommandAsDifferentUser struct {
	mu sync.RWMutex
	// changeUserCmd describes the command to use to execute any command as a different user
	// e.g. it can be set as "sudo" to run commands as `root` or as "su","tom" or "gosu","jack"
	changeUserCmd []string
	prepend       *CommandAsDifferentUser
}

// Redefine redefines a command so that it will be run as a different user.
func (c *CommandAsDifferentUser) Redefine(cmd string, args ...string) (cmdName string, cmdArgs []string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	newArgs := []string{cmd}
	newArgs = append(newArgs, args...)
	cmdName, cmdArgs = c.redefineCommand(newArgs...)
	return
}

// RedefineCommand is the same as Redefine but with no separation between the command and its arguments (like the command in Docker)
func (c *CommandAsDifferentUser) RedefineCommand(args ...string) (cmdName string, cmdArgs []string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cmdName, cmdArgs = c.redefineCommand(args...)
	return
}

func (c *CommandAsDifferentUser) redefineCommand(args ...string) (cmdName string, cmdArgs []string) {
	if len(c.changeUserCmd) > 0 {
		cmdName = c.changeUserCmd[0]
		cmdArgs = append(cmdArgs, c.changeUserCmd[1:]...)
		cmdArgs = append(cmdArgs, args...)
	} else {
		cmdName = args[0]
		cmdArgs = append(cmdArgs, args[1:]...)
	}
	if c.prepend != nil {
		cmdName, cmdArgs = c.prepend.Redefine(cmdName, cmdArgs...)
		return
	}
	return
}

// RedefineInShellForm returns the new command defined in shell form.
func (c *CommandAsDifferentUser) RedefineInShellForm(cmd string, args ...string) string {
	ncmd, nargs := c.Redefine(cmd, args...)
	return AsShellForm(ncmd, nargs...)
}

// Prepend prepends a command translator to this command translator
// This can be used to run command as a separate user and with higher privileges e.g. `sudo`.
// It returns this for use in fluent expressions e.g. Gosu(...).Prepend(Sudo()).RedefineInShellForm(...)
func (c *CommandAsDifferentUser) Prepend(cmd *CommandAsDifferentUser) *CommandAsDifferentUser {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prepend = cmd
	return c
}

// NewCommandAsDifferentUser defines a command wrapper which helps to redefine commands so that they are run as a different user.
// e.g.
//   - switchUserCmd="sudo" to run commands as `root`
//   - switchUserCmd="su", "tom" if `tom` has enough privileges to run the command
//   - switchUserCmd="gosu", "tom" if `tom` has enough privileges to run the command in a container and `gosu` is installed
func NewCommandAsDifferentUser(switchUserCmd ...string) *CommandAsDifferentUser {
	return &CommandAsDifferentUser{changeUserCmd: switchUserCmd}
}

// NewCommandAsRoot will create a command translator which will run command with `sudo` (for Unix Only)
func NewCommandAsRoot() *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("sudo")
}

// Sudo will call commands with `sudo`. Similar to NewCommandAsRoot (for Unix Only)
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

// Su will run commands as the user username using [su](https://www.unix.com/man-page/posix/1/su/)  (for Unix Only)
func Su(username string) *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("su", username)
}

// Me will run the commands without switching user. It is a no operation wrapper.
func Me() *CommandAsDifferentUser {
	return NewCommandAsDifferentUser()
}

// RunAs will run commands as the user username using [runas](https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2012-r2-and-2012/cc771525%28v=ws.11%29) (For WINDOWS only)
func RunAs(username string) *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("runas", fmt.Sprintf("/user:%v", username))
}

// Elevate will call commands with [elevate](https://learn.microsoft.com/en-us/previous-versions/technet-magazine/cc162321(v=msdn.10)) assuming the tool is installed on the platform. (For WINDOWS only)
func Elevate() *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("elevate")
}

// ShellRunAs will call commands with [shellrunas](https://learn.microsoft.com/en-gb/sysinternals/downloads/shellrunas) assuming [SysInternals](https://docs.microsoft.com/en-gb/sysinternals/downloads/sysinternals-suite) is installed on the platform. (For WINDOWS only)
func ShellRunAs() *CommandAsDifferentUser {
	return NewCommandAsDifferentUser("shellrunas", "/quiet")
}

// AsShellForm returns a command in its shell form.
func AsShellForm(cmd string, args ...string) string {
	newCmd := []string{cmd}
	newCmd = append(newCmd, args...)
	return strings.Join(newCmd, " ")
}
