/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package platform

import (
	"os/user"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

// IsCurrentWithRootPrivileges reports whether the current process is running
// with elevated privileges.
func IsCurrentWithRootPrivileges() (bool, error) {
	return isCurrentWithRootPrivileges()
}

// CanCurrentUserMutateUsers reports whether the current process has enough
// privileges and platform support to create, modify, and remove local users
// and groups.
func CanCurrentUserMutateUsers() (bool, error) {
	return canCurrentUserMutateUsersOnPlatform()
}

// IsCurrentUserAnAdmin states whether the current user is a superuser or not.
func IsCurrentUserAnAdmin() (admin bool, err error) {
	cuser, err := user.Current()
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "cannot fetch the current user")
		return
	}
	admin, err = IsUserAdmin(cuser)
	return
}

// IsUserAdmin states whether the user is a superuser or not. Similar to IsAdmin but may use more checks.
func IsUserAdmin(user *user.User) (admin bool, err error) {
	if user == nil {
		err = commonerrors.UndefinedVariable("user")
		return
	}
	admin, subErr := isUserAdmin(user)
	if subErr == nil {
		return
	}
	admin, err = IsAdmin(user.Username)
	err = ConvertUserGroupError(err)
	return
}

// IsAdmin states whether the user is a superuser or not.
func IsAdmin(username string) (admin bool, err error) {
	found, subErr := HasUser(username)
	if !found && subErr == nil {
		return
	}
	admin, err = isAdmin(username)
	err = ConvertUserGroupError(err)
	if err == nil {
		return
	}
	// Make more check if username is current user
	current, subErr := user.Current()
	if subErr != nil {
		return
	}
	if current.Username != username {
		return
	}
	admin, err = isCurrentAdmin()
	err = ConvertUserGroupError(err)
	return
}

// WithPrivileges redefines a command so that it is run with elevated privileges.
// For instance, on Linux, if the current user has enough privileges, the command will be run as is.
// Otherwise, `sudo` will be used if defined as the sudo  (See `DefineSudoCommand`).
// Similar scenario will happen on Windows, although the elevated command is defined using `DefineSudoCommand`.
func WithPrivileges(cmd *command.CommandAsDifferentUser) (cmdWithPrivileges *command.CommandAsDifferentUser) {
	cmdWithPrivileges = cmd
	if cmdWithPrivileges == nil {
		cmdWithPrivileges = command.NewCommandAsDifferentUser()
	}
	hasPrivileges, err := IsCurrentUserAnAdmin()
	if err != nil {
		return
	}
	if !hasPrivileges && hasPasswordlessPrivileges() {
		cmdWithPrivileges.Prepend(getRunCommandWithPrivileges())
	}
	return
}
