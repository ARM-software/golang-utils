//go:build linux || unix || (js && wasm) || darwin || aix || dragonfly || freebsd || nacl || netbsd || openbsd || solaris

package platform

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

const (
	RootGroup    = "root"
	SudoersGroup = "sudo"
)

func addUser(ctx context.Context, username, fullname, password string) (err error) {
	pwd := password
	if pwd == "" {
		pwd = "tmp123"
	}
	comment := fullname
	if comment == "" {
		comment = username
	}
	err = executeCommand(ctx, "useradd", "-U", "-m", username, "-p", pwd, "-c", comment)
	if err != nil {
		return
	}
	err = executeCommand(ctx, "passwd", "-d", username)
	return

}

func removeUser(ctx context.Context, username string) (err error) {
	err = executeCommand(ctx, "userdel", "-f", "-r", username)
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	err = executeCommand(ctx, "groupadd", "-f", groupName)
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	err = executeCommand(ctx, "groupdel", groupName)
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	err = executeCommand(ctx, "usermod", "-a", "-G", groupName, username)
	return
}

func dissociateUserFromGroup(ctx context.Context, username, groupName string) (err error) {
	err = executeCommand(ctx, "gpasswd", "-d", username, groupName)
	return
}

func isAdmin(username string) (admin bool, err error) {
	if username == "root" {
		admin = true
		return
	}
	err = fmt.Errorf("%w: could not check whether the user is a superuser or not", commonerrors.ErrNotImplemented)
	return
}

func isUserAdmin(user *user.User) (admin bool, err error) {
	// following method mentioned [here](https://linuxhandbook.com/check-if-user-has-sudo-rights/#method-2-check-if-user-is-part-of-the-sudo-group)
	gids, subErr := user.GroupIds()

	if subErr == nil {
		_, admin = collection.FindInSlice(true, gids, RootGroup, SudoersGroup)
		if admin {
			return
		}
	}
	admin, err = isAdmin(user.Username)
	return
}

func isCurrentAdmin() (admin bool, err error) {
	// It is not great but following [this way](https://serverfault.com/questions/568627/can-a-program-tell-it-is-being-run-under-sudo)
	// also mentioned [here](https://stackoverflow.com/questions/29733575/how-to-find-the-user-that-executed-a-program-as-root-using-golang)
	// and [here](https://stackoverflow.com/questions/10272784/how-do-i-get-the-users-real-uid-if-the-program-is-run-with-sudo)
	if collection.AllNotEmpty(true, []string{
		os.Getenv("SUDO_USER"), os.Getenv("SUDO_COMMAND"), os.Getenv("SUDO_UID"), os.Getenv("SUDO_GID"),
	}) {
		admin = true
	}
	return
}
