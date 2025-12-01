//go:build windows

package platform

import (
	"context"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// Note : there is no implementation for the arm architecture until support is provided by the following library https://github.com/iamacarpet/go-win64api/issues/62

func addUser(ctx context.Context, username, fullname, password string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fmt.Errorf("%w: cannot add user", commonerrors.ErrNotImplemented)
	return
}

func removeUser(ctx context.Context, username string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fmt.Errorf("%w: cannot remove user", commonerrors.ErrNotImplemented)
	return
}

func addGroup(ctx context.Context, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fmt.Errorf("%w: cannot add group", commonerrors.ErrNotImplemented)
	return
}

func associateUserToGroup(ctx context.Context, username, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fmt.Errorf("%w: cannot associate user to group", commonerrors.ErrNotImplemented)
	return
}

func dissociateUserFromGroup(ctx context.Context, username, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fmt.Errorf("%w: cannot dissociate user from group", commonerrors.ErrNotImplemented)
	return
}

func removeGroup(ctx context.Context, groupName string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	err = fmt.Errorf("%w: cannot associate user to group", commonerrors.ErrNotImplemented)
	return
}

func isAdmin(username string) (admin bool, err error) {
	err = fmt.Errorf("%w: cannot determine if user is admin", commonerrors.ErrNotImplemented)
	return
}
