//go:build linux

package subprocess

import (
	"os/exec"
	"syscall"
)

// See https://github.com/tgulacsi/go/blob/master/proc/
func setGroupAttrToCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true, // to be able to kill all children, too
		Pdeathsig: syscall.SIGKILL,
	}
}
