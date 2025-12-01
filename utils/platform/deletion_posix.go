//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris

package platform

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

func removeFileAs(ctx context.Context, as *command.CommandAsDifferentUser, path string) error {
	return executeCommandAs(ctx, as, "rm", "-f")
}

func removeDirAs(ctx context.Context, as *command.CommandAsDifferentUser, path string) error {
	return executeCommandAs(ctx, as, "rm", "-r", "-f")
}
