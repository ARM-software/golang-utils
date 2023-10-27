package platform

import (
	"fmt"

	"github.com/mitchellh/go-homedir"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// GetHomeDirectory returns the home directory of a user.
func GetHomeDirectory(username string) (string, error) {
	user, err := GetUser(username)
	if err != nil {
		return GetDefaultHomeDirectory(username)
	}
	return fetchHomeDirectory(username, user.HomeDir)
}

// GetDefaultHomeDirectory returns the default home directory for a user based on the convention listed in https://en.wikipedia.org/wiki/Home_directory#Default_home_directory_per_operating_system
func GetDefaultHomeDirectory(username string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("%w: missing username", commonerrors.ErrUndefined)
	}
	return determineDefaultHomeDirectory(username)
}

func fetchHomeDirectory(username, homedirPath string) (home string, err error) {
	home, err = homedir.Expand(homedirPath)
	if err == nil {
		return
	}
	home, err = GetDefaultHomeDirectory(username)
	if err != nil {
		err = fmt.Errorf("%w: could not determine user [%v]'s home directory: %v", commonerrors.ErrUnexpected, username, err.Error())
	}
	return
}
