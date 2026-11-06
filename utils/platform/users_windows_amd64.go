/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build windows

package platform

import (
	"context"

	wapi "github.com/iamacarpet/go-win64api"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

func addUser(ctx context.Context, username, fullname, password string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	success, err := wapi.UserAdd(username, fullname, password)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not add user")
		return
	}
	if !success {
		err = commonerrors.New(commonerrors.ErrUnknown, "failed adding user")
		return
	}
	return
}

func removeUser(ctx context.Context, username string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	success, err := wapi.UserDelete(username)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not remove user")
		return
	}
	if !success {
		err = commonerrors.New(commonerrors.ErrUnknown, "failed removing user")
		return
	}
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	success, err := wapi.LocalGroupAdd(groupName, "new group")
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not add group")
		return
	}
	if !success {
		err = commonerrors.New(commonerrors.ErrUnknown, "failed adding group")
		return
	}
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	success, err := wapi.AddGroupMembership(username, groupName)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not associate user to group")
		return
	}
	if !success {
		err = commonerrors.New(commonerrors.ErrUnknown, "failed associating user to group")
		return
	}
	return
}

func dissociateUserFromGroup(ctx context.Context, username, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	success, err := wapi.RemoveGroupMembership(username, groupName)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not dissociate user to group")
		return
	}
	if !success {
		err = commonerrors.New(commonerrors.ErrUnknown, "failed dissociating user to group")
		return
	}
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	success, err := wapi.LocalGroupDel(groupName)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not remove group")
		return
	}
	if !success {
		err = commonerrors.New(commonerrors.ErrUnknown, "failed removing group")
		return
	}
	return
}

func isAdmin(username string) (admin bool, err error) {
	admin, err = wapi.IsLocalUserAdmin(username)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not determine if user is an admin")
	}
	return
}
