//go:build unix || (js && wasm)
// +build unix js,wasm

package filesystem

import (
	"os"
	"syscall"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func determineFileOwners(info os.FileInfo) (uid, gid int, err error) {
	if raw, ok := info.Sys().(*syscall.Stat_t); ok {
		gid = int(raw.Gid)
		uid = int(raw.Uid)
	} else {
		err = commonerrors.ErrNotImplemented
	}
	return
}
