package proc

import "os/exec"

// See https://github.com/tgulacsi/go/blob/master/proc/
func SetGroupAttrToCmd(c *exec.Cmd) {
	setGroupAttrToCmd(c)
}
