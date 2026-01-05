//go:build darwin

package proc

import (
	"os/exec"
	"syscall"
)

// See https://github.com/tgulacsi/go/blob/master/proc/proc_darwin.go
func setGroupAttrToCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
