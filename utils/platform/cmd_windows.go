//go:build windows

package platform

import (
	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

var (
	// runAsAdmin describes the command to use to run command as Administrator
	// See https://ss64.com/nt/syntax-elevate.html
	// see https://lazyadmin.nl/it/runas-command/
	// see https://www.tenforums.com/general-support/111929-how-use-runas-without-password-prompt.html
	runAsAdmin = command.RunAs("Administrator")
)

// DefineRunAsAdmin defines the command to run as Administrator.
// e.g.
//   - args="runas", "/user:adrien" to run commands as `adrien`
func DefineRunAsAdmin(args ...string) {
	runAsAdmin = command.NewCommandAsDifferentUser(args...)
}

func getRunCommandWithPrivileges() *command.CommandAsDifferentUser {
	return runAsAdmin
}

func hasPasswordlessPrivileges() bool {
	return true
}
