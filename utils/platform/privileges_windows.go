//go:build windows

/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/subprocess/command"
)

var (
	// runAsAdmin describes the command to use to run command as Administrator
	// See https://ss64.com/nt/syntax-elevate.html
	// see https://lazyadmin.nl/it/runas-command/
	// see https://www.tenforums.com/general-support/111929-how-use-runas-without-password-prompt.html
	runAsAdmin = command.RunAs("Administrator")
)

// DefineRunAsAdmin defines the command to run as Administrator.
// e.g.
//   - args="runas", "/user:adrien" to run commands as `adrien`
func DefineRunAsAdmin(args ...string) {
	runAsAdmin = command.NewCommandAsDifferentUser(args...)
}

func isCurrentWithRootPrivileges() (bool, error) {
	return IsCurrentUserAnAdmin()
}

func canCurrentUserMutateUsersOnPlatform() (bool, error) {
	admin, err := IsCurrentUserAnAdmin()
	if err != nil || !admin {
		return admin, err
	}

	// A positive admin check is not enough on hosted runners: local account and
	// group management can still fail because the current token cannot update the
	// local SAM database. Probe a disposable local group to confirm the runner
	// can actually perform user/group mutation operations.
	// See:
	// - https://learn.microsoft.com/windows/win32/api/lmaccess/nf-lmaccess-netlocalgroupadd
	// - https://learn.microsoft.com/windows/win32/netmgmt/network-management-error-codes
	probeGroup := fmt.Sprintf("gutest_%d", time.Now().UnixNano())
	err = AddGroup(context.Background(), probeGroup)
	if err != nil {
		// `status=5` corresponds to access denied from the underlying Windows
		// network-management APIs when the runner is effectively unable to manage
		// local groups despite reporting administrative membership.
		if commonerrors.CorrespondTo(err, "status=5", "access is denied") {
			return false, nil
		}
		return false, err
	}
	defer func() { _ = RemoveGroup(context.Background(), probeGroup) }()
	return true, nil
}

func getRunCommandWithPrivileges() *command.CommandAsDifferentUser {
	return runAsAdmin
}

func hasPasswordlessPrivileges() bool {
	return true
}
