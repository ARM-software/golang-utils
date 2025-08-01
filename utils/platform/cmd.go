package platform

import "github.com/ARM-software/golang-utils/utils/subprocess/command"

// WithPrivileges redefines a command so that it is run with elevated privileges.
// For instance, on Linux, if the current user has enough privileges, the command will be run as is.
// Otherwise, `sudo` will be used if defined as the sudo  (See `DefineSudoCommand`).
// Similar scenario will happen on Windows, although the elevated command is defined using `DefineSudoCommand`.
func WithPrivileges(cmd *command.CommandAsDifferentUser) (cmdWithPrivileges *command.CommandAsDifferentUser) {
	cmdWithPrivileges = cmd
	if cmdWithPrivileges == nil {
		cmdWithPrivileges = command.NewCommandAsDifferentUser()
	}
	hasPrivileges, err := IsCurrentUserAnAdmin()
	if err != nil {
		return
	}
	if !hasPrivileges && hasPasswordlessPrivileges() {
		cmdWithPrivileges.Prepend(getRunCommandWithPrivileges())
	}
	return
}
