//go:build windows

package platform

import (
	"os"
	"os/exec"
	"os/user"

	"golang.org/x/sys/windows"

	"github.com/ARM-software/golang-utils/utils/collection"
)

const (
	AdministratorsGroup = "S-1-5-32-544" //https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers
)

func isUserAdmin(user *user.User) (admin bool, err error) {
	gids, subErr := user.GroupIds()
	if subErr == nil {
		_, admin = collection.FindInSlice(true, gids, AdministratorsGroup)
		if admin {
			return
		}
	}
	admin, err = isAdmin(user.Username)
	return
}

func isCurrentAdmin() (admin bool, err error) {
	// In order to avoid, this https://stackoverflow.com/questions/8046097/how-to-check-if-a-process-has-the-administrative-rights
	// https://gist.github.com/jerblack/d0eb182cc5a1c1d92d92a4c4fcc416c6
	file, subErr := os.Open("\\\\.\\PHYSICALDRIVE0")
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	admin = subErr == nil
	if admin {
		return
	}
	// extra check https://stackoverflow.com/a/28268802
	subErr = exec.Command("fltmc").Run()
	admin = subErr == nil
	if admin {
		return
	}
	admin = windows.GetCurrentProcessToken().IsElevated()
	return
}
