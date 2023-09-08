package environment

import (
	"os"
	"os/user"

	"github.com/mitchellh/go-homedir"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

type currentEnv struct {
}

func (c *currentEnv) GetCurrentUser() (currentUser *user.User) {
	currentUser, err := user.Current()
	if err != nil {
		return
	}
	home, err := homedir.Dir()
	if err != nil {
		return
	}
	currentUser.HomeDir = home
	return
}

func (c *currentEnv) GetEnvironmentVariables() (variables []IEnvironmentVariable) {
	curentEnv := os.Environ()
	for i := range curentEnv {
		envvar, err := ParseEnvironmentVariable(curentEnv[i])
		if err != nil {
			return
		}
		variables = append(variables, envvar)
	}
	return
}

func (c *currentEnv) GetFilesystem() filesystem.FS {
	return filesystem.NewStandardFileSystem()
}

// NewCurrentEnvironment returns system current environment.
func NewCurrentEnvironment() IEnvironment {
	return &currentEnv{}
}
