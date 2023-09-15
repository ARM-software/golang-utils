package platform

import (
	"context"
	"fmt"
	"os/user"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// DefineUser adds a new user to the platform
func DefineUser(ctx context.Context, user *user.User, password string) (err error) {
	if user == nil {
		err = fmt.Errorf("%w: missing user information", commonerrors.ErrUndefined)
		return
	}
	username := user.Username
	if username == "" {
		username = user.Uid
	}
	err = AddUser(ctx, user.Username, user.Name, password)
	return
}

// AddUser adds a new user to the platform
func AddUser(ctx context.Context, username, fullname, password string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	found, _ := HasUser(username)
	if found {
		return nil
	}
	return ConvertUserError(addUser(ctx, username, fullname, password))
}

// DeleteUser removes a user from the platform
func DeleteUser(ctx context.Context, user *user.User) (err error) {
	if user == nil {
		return
	}
	username := user.Username
	if username == "" {
		username = user.Uid
	}
	err = RemoveUser(ctx, username)
	return
}

// RemoveUser removes a user from the platform
func RemoveUser(ctx context.Context, username string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	return ConvertUserError(removeUser(ctx, username))
}

// HasUser checks whether a user exists
func HasUser(username string) (found bool, err error) {
	user, err := user.Lookup(username)
	if err != nil {
		err = ConvertUserError(err)
		if commonerrors.Any(err, commonerrors.ErrNotFound) {
			err = nil
		}
	}
	if user != nil {
		found = true
	}
	return
}

// HasGroup checks whether a group exists
func HasGroup(groupName string) (found bool, err error) {
	group, err := user.LookupGroup(groupName)
	if err != nil {
		err = ConvertUserError(err)
		if commonerrors.Any(err, commonerrors.ErrNotFound) {
			err = nil
		}
	}
	if group != nil {
		found = true
	}
	return
}

// ConvertUserError converts errors related to users in common errors.
func ConvertUserError(err error) error {
	if err == nil {
		return nil
	}
	err = commonerrors.ConvertContextError(err)
	switch {
	case commonerrors.Any(err, commonerrors.ErrTimeout, commonerrors.ErrCancelled, commonerrors.ErrUnknown, commonerrors.ErrUnexpected, commonerrors.ErrNotFound):
		return err
	case commonerrors.CorrespondTo(err, "unknown"):
		return fmt.Errorf("%w: %v", commonerrors.ErrNotFound, err.Error())
	}
	return fmt.Errorf("%w: %v", commonerrors.ErrUnknown, err.Error())
}
