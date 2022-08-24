package filesystem

import (
	"os"
	"syscall"
)

func determineFileOwners(os.FileInfo) (uid, gid int, err error) {
	uid = syscall.Getuid()
	gid = syscall.Getgid()

	return
}
