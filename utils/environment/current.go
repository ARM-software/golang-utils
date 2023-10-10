package environment

import (
	"fmt"
	"os"
	"os/user"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
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

// GetEnvironmentVariables returns the current environment variable (and optionally those in the supplied dotEnvFiles)
func (c *currentEnv) GetEnvironmentVariables(dotEnvFiles ...string) (variables []IEnvironmentVariable) {
	if len(dotEnvFiles) > 0 { // if no args, then it will attempt to load .env
		_ = godotenv.Load(dotEnvFiles...) // ignore error (specifically on loading .env) consistent with config.LoadFromEnvironment
	}

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

// GetEnvironmentVariable searchs the current environment (and optionally dotEnvFiles) for a specific env var
func (c *currentEnv) GetEnvironmentVariable(envvar string, dotEnvFiles ...string) (value IEnvironmentVariable, err error) {
	envvars := c.GetEnvironmentVariables(dotEnvFiles...)
	for i := range envvars {
		if envvars[i].GetKey() == envvar {
			return envvars[i], nil
		}
	}
	return nil, fmt.Errorf("%w: environment variable '%v' not set", commonerrors.ErrNotFound, envvar)
}

// NewCurrentEnvironment returns system current environment.
func NewCurrentEnvironment() IEnvironment {
	return &currentEnv{}
}
