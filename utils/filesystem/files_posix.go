//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm darwin aix dragonfly freebsd nacl netbsd openbsd solaris

package filesystem

import (
	"os"
	"syscall"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func isPrivilegeError(err error) bool {
	return commonerrors.Any(err, syscall.EPERM)
}

func isNotSupportedError(err error) bool {
	return commonerrors.Any(err, syscall.ENOTSUP, syscall.EOPNOTSUPP)
}

func determineFileOwners(info os.FileInfo) (uid, gid int, err error) {
	if raw, ok := info.Sys().(*syscall.Stat_t); ok {
		gid = int(raw.Gid)
		uid = int(raw.Uid)
	} else {
		err = commonerrors.Newf(commonerrors.ErrUnsupported, "file info [%v] is not of type Stat_t", info.Sys())
	}
	return
}
