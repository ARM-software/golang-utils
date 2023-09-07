// Package environment defines object describing the current environment.
package environment

import (
	"encoding"
	"fmt"
	"os/user"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IEnvironmentVariable,IEnvironment

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
	// GetEnvironmentVariables returns the variables defining the environment.
	GetEnvironmentVariables() []IEnvironmentVariable
	// GetFilesystem returns the filesystem associated with the current environment
	GetFilesystem() filesystem.FS
}
