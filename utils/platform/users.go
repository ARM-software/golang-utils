package platform

import (
	"context"
	"fmt"
	"os/user"

	"github.com/mitchellh/go-homedir"

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
	err = AddUser(ctx, username, user.Name, password)
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
	return ConvertUserGroupError(addUser(ctx, username, fullname, password))
}

// DeleteUser removes a user from the platform when the user is specified using a `user.User` structure.
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

// RemoveUser removes a user from the platform when only the username is known.
func RemoveUser(ctx context.Context, username string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	return ConvertUserGroupError(removeUser(ctx, username))
}

// HasUser checks whether a user exists
func HasUser(username string) (found bool, err error) {
	user, err := user.Lookup(username)
	if err != nil {
		err = ConvertUserGroupError(err)
		if commonerrors.Any(err, commonerrors.ErrNotFound) {
			err = nil
		}
	}
	if user != nil {
		found = true
	}
	return
}

// GetUser returns information about a user and expands its home directory.
func GetUser(username string) (auser *user.User, err error) {
	auser, err = user.Lookup(username)
	if err != nil {
		err = ConvertUserGroupError(err)
		return
	}
	if auser == nil {
		err = fmt.Errorf("%w: missing user information", commonerrors.ErrUnexpected)
		return
	}
	home, err := fetchHomeDirectory(username, auser.HomeDir)
	if err != nil {
		return
	}
	auser.HomeDir = home
	return
}

// GetCurrentUser returns information about the current platform's user and expands its home directory.
func GetCurrentUser() (currentUser *user.User, err error) {
	currentUser, err = user.Current()
	if err != nil {
		err = ConvertUserGroupError(err)
		return
	}
	if currentUser == nil {
		err = fmt.Errorf("%w: missing user information", commonerrors.ErrUnexpected)
		return
	}
	home, err := homedir.Dir()
	if err != nil {
		err = fmt.Errorf("%w: could not retrieve information about current user's home directory: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	currentUser.HomeDir = home
	return
}

// IsCurrentUserAnAdmin states whether the current user is a superuser or not.
func IsCurrentUserAnAdmin() (admin bool, err error) {
	cuser, err := user.Current()
	if err != nil {
		err = fmt.Errorf("%w: cannot fetch the current user: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}
	admin, err = IsUserAdmin(cuser)
	return
}

// IsUserAdmin states whether the user is a superuser or not. Similar to IsAdmin but may use more checks.
func IsUserAdmin(user *user.User) (admin bool, err error) {
	if user == nil {
		err = fmt.Errorf("%w: missing user", commonerrors.ErrUndefined)
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

// HasGroup checks whether a group exists
func HasGroup(groupName string) (found bool, err error) {
	group, err := user.LookupGroup(groupName)
	if err != nil {
		err = ConvertUserGroupError(err)
		if commonerrors.Any(err, commonerrors.ErrNotFound) {
			err = nil
		}
	}
	if group != nil {
		found = true
	}
	return
}

// AddGroup creates a group if not already existing.
func AddGroup(ctx context.Context, groupName string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	found, _ := HasGroup(groupName)
	if found {
		return nil
	}
	return ConvertUserGroupError(addGroup(ctx, groupName))
}

// RemoveGroup removes a group from the platform
func RemoveGroup(ctx context.Context, groupName string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	return ConvertUserGroupError(removeGroup(ctx, groupName))
}

// AssociateUserToGroup adds a user to a group.
func AssociateUserToGroup(ctx context.Context, username, groupName string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	found, err := HasGroup(groupName)
	if err != nil || !found {
		return fmt.Errorf("%w: the group does not seem to exist: %v", commonerrors.ErrNotFound, err.Error())
	}
	found, err = HasUser(username)
	if err != nil || !found {
		return fmt.Errorf("%w: the user does not seem to exist: %v", commonerrors.ErrNotFound, err.Error())
	}
	return ConvertUserGroupError(associateUserToGroup(ctx, username, groupName))
}

// DissociateUserFromGroup removes a user from a group.
func DissociateUserFromGroup(ctx context.Context, username, groupName string) error {
	err := parallelisation.DetermineContextError(ctx)
	if err != nil {
		return err
	}
	found, _ := HasGroup(groupName)
	if !found && err == nil {
		return nil
	}
	found, _ = HasUser(username)
	if !found && err == nil {
		return nil
	}
	return ConvertUserGroupError(dissociateUserFromGroup(ctx, username, groupName))
}

// ConvertUserGroupError converts errors related to users in common errors.
func ConvertUserGroupError(err error) error {
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
