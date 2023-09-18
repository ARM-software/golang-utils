//go:build windows
// +build windows

// Package filesystem describes the filesystem on windows
package filesystem

import (
	"os"
	"syscall"
)

func determineFileOwners(_ os.FileInfo) (uid, gid int, err error) {
	uid = syscall.Getuid()
	gid = syscall.Getgid()

	return
}
