//go:build windows
// +build windows

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
