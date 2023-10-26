package environment

import (
	"os"
	"os/user"

	"github.com/joho/godotenv"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/platform"
)

type currentEnv struct {
}

func (c *currentEnv) GetCurrentUser() (currentUser *user.User) {
	currentUser, _ = platform.GetCurrentUser()
	return
}

// GetEnvironmentVariables returns the current environment variable (and optionally those in the supplied in `dotEnvFiles`)
// `dotEnvFiles` corresponds to `.env` files present on the machine and follows the mechanism described by https://github.com/bkeepers/dotenv
func (c *currentEnv) GetEnvironmentVariables(dotEnvFiles ...string) (variables []IEnvironmentVariable) {
	if len(dotEnvFiles) > 0 { // if no args, then it will attempt to load .env
		_ = godotenv.Load(dotEnvFiles...) // ignore error (specifically on loading .env) consistent with config.LoadFromEnvironment
	}

	variables = ParseEnvironmentVariables(os.Environ()...)
	return
}

func (c *currentEnv) GetFilesystem() filesystem.FS {
	return filesystem.NewStandardFileSystem()
}

// GetEnvironmentVariable searches the current environment (and optionally dotEnvFiles) for a specific environment variable `envvar`.
func (c *currentEnv) GetEnvironmentVariable(envvar string, dotEnvFiles ...string) (value IEnvironmentVariable, err error) {
	return FindEnvironmentVariable(envvar, c.GetEnvironmentVariables(dotEnvFiles...)...)
}

// NewCurrentEnvironment returns system current environment.
func NewCurrentEnvironment() IEnvironment {
	return &currentEnv{}
}
