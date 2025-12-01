//go:build windows

package platform

import (
	"context"
	"fmt"

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
		err = fmt.Errorf("%w: could not add user: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	if !success {
		err = fmt.Errorf("%w: failed adding user", commonerrors.ErrUnknown)
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
		err = fmt.Errorf("%w: could not remove user: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	if !success {
		err = fmt.Errorf("%w: failed removing user", commonerrors.ErrUnknown)
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
		err = fmt.Errorf("%w: could not add group: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	if !success {
		err = fmt.Errorf("%w: failed adding group", commonerrors.ErrUnknown)
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
		err = fmt.Errorf("%w: could not associate user to group: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	if !success {
		err = fmt.Errorf("%w: failed associating user to group", commonerrors.ErrUnknown)
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
		err = fmt.Errorf("%w: could not dissociate user to group: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	if !success {
		err = fmt.Errorf("%w: failed dissociating user to group", commonerrors.ErrUnknown)
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
		err = fmt.Errorf("%w: could not remove group: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	if !success {
		err = fmt.Errorf("%w: failed removing group", commonerrors.ErrUnknown)
		return
	}
	return
}

func isAdmin(username string) (admin bool, err error) {
	admin, err = wapi.IsLocalUserAdmin(username)
	if err != nil {
		err = fmt.Errorf("%w: could not determine if user is an admin: %v", commonerrors.ErrUnexpected, err.Error())
	}
	return
}
