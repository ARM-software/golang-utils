/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//go:build windows

package platform

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// Note : there is no implementation for the arm architecture until support is provided by the following library https://github.com/iamacarpet/go-win64api/issues/62

func addUser(ctx context.Context, username, fullname, password string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = commonerrors.New(commonerrors.ErrNotImplemented, "cannot add user")
	return
}

func removeUser(ctx context.Context, username string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = commonerrors.New(commonerrors.ErrNotImplemented, "cannot remove user")
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = commonerrors.New(commonerrors.ErrNotImplemented, "cannot add group")
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = commonerrors.New(commonerrors.ErrNotImplemented, "cannot associate user to group")
	return
}

func dissociateUserFromGroup(ctx context.Context, username, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = commonerrors.New(commonerrors.ErrNotImplemented, "cannot dissociate user from group")
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = commonerrors.New(commonerrors.ErrNotImplemented, "cannot associate user to group")
	return
}

func isAdmin(username string) (admin bool, err error) {
	err = commonerrors.New(commonerrors.ErrNotImplemented, "cannot determine if user is admin")
	return
}
