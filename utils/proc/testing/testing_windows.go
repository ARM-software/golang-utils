//go:build windows

package testing

import (
	"os/exec"
	"syscall"
)

func SetGroupAttrToCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		// for any of our checks to work we need to set the group ID to the pid, otherwise the group ID
		// will be the code that launched it (e.g. the exec in the test). This causes issues in tests as
		// any checks for running processes will return the test PID not the sub process one.
		HideWindow: true,
		// Windows Process Creation Flags:  https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
		CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT | syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
