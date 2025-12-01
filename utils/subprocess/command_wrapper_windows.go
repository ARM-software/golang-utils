//go:build windows

package subprocess

import (
	"os/exec"
	"syscall"
)

func setGroupAttrToCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		// Windows Process Creation Flags:  https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
		CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT | syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
