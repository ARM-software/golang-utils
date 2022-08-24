//go:build unix || (js && wasm) || darwin
// +build unix js,wasm darwin

package filesystem

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func determineFileOwners(info os.FileInfo) (uid, gid int, err error) {
	if raw, ok := info.Sys().(*syscall.Stat_t); ok {
		gid = int(raw.Gid)
		uid = int(raw.Uid)
	} else {
		err = fmt.Errorf("%w: file info [%v] is not of type Stat_t", commonerrors.ErrUnsupported, info.Sys())
	}
	return
}
