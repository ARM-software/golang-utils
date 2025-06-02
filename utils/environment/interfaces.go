// Package environment defines object describing the current environment.
package environment

import (
	"encoding"
	"fmt"
	"os/user"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IEnvironmentVariable,IEnvironment

// IEnvironmentVariable defines an environment variable to be set for the commands to run.
type IEnvironmentVariable interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	fmt.Stringer
	// GetKey returns the variable key.
	GetKey() string
	// GetValue returns the variable value.
	GetValue() string
	// Validate checks whether the variable value is correctly defined
	Validate() error
	// Equal states whether two environment variables are equal or not.
	Equal(v IEnvironmentVariable) bool
}

// IEnvironment defines an environment for an application to run on.
type IEnvironment interface {
	// GetCurrentUser returns the environment current user.
	GetCurrentUser() *user.User
	// GetEnvironmentVariables returns the variables defining the environment  (and optionally those supplied in `dotEnvFiles`)
	// `dotEnvFiles` corresponds to `.env` files present on the machine and follows the mechanism described by https://github.com/bkeepers/dotenv
	GetEnvironmentVariables(dotEnvFiles ...string) []IEnvironmentVariable
	// GetExpandedEnvironmentVariables  is similar to GetEnvironmentVariables but returns variables with fully expanded values.
	// e.g. on Linux, if variable1=${variable2}, then the reported value of variable1 will be the value of variable2
	GetExpandedEnvironmentVariables(dotEnvFiles ...string) []IEnvironmentVariable
	// GetFilesystem returns the filesystem associated with the current environment
	GetFilesystem() filesystem.FS
	// GetEnvironmentVariable returns the environment variable corresponding to `envvar` or an error if it not set. optionally it searches `dotEnvFiles` files too
	// `dotEnvFiles` corresponds to `.env` files present on the machine and follows the mechanism described by https://github.com/bkeepers/dotenv
	GetEnvironmentVariable(envvar string, dotEnvFiles ...string) (IEnvironmentVariable, error)
	// GetExpandedEnvironmentVariable is similar to GetEnvironmentVariable but returns variables with fully expanded values.
	//	// e.g. on Linux, if variable1=${variable2}, then the reported value of variable1 will be the value of variable2
	GetExpandedEnvironmentVariable(envvar string, dotEnvFiles ...string) (IEnvironmentVariable, error)
}
