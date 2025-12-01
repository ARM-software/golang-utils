//go:build !darwin && !windows

package platform

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/mitchellh/go-homedir"
)

func determineDefaultHomeDirectory(username string) (string, error) {
	currentDir, subErr1 := homedir.Dir()
	currentUser, subErr2 := user.Current()
	if subErr1 != nil || subErr2 != nil {
		return fmt.Sprintf("/home/%v", username), nil
	}
	return strings.ReplaceAll(currentDir, currentUser.Username, username), nil
}
