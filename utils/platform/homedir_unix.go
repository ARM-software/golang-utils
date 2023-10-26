//go:build linux || unix || (js && wasm) || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris
// +build linux unix js,wasm aix dragonfly freebsd nacl netbsd openbsd solaris

package platform

import (
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

func determineDefaultHomeDirectory(username string) (string, error) {
	currentDir, subErr1 := homedir.Dir()
	currentUser, subErr2 := user.Current()
	if subErr1 != nil || subErr2 != nil {
		return filepath.Join(`/`, "home", username)
	}
	return strings.ReplaceAll(currentDir, currentUser.Username, username), nil
}
