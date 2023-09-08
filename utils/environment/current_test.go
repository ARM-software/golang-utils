package environment

import (
	"os"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCurrentEnvironment(t *testing.T) {
	current := NewCurrentEnvironment()
	currentUser := current.GetCurrentUser()
	require.NotNil(t, currentUser)
	require.NotEmpty(t, currentUser.HomeDir)
	require.NotEmpty(t, currentUser.Username)
	currentFs := current.GetFilesystem()
	require.NotNil(t, currentFs)
}

func Test_currentEnv_GetEnvironmentVariables(t *testing.T) {
	os.Clearenv()
	require.NoError(t, os.Setenv("test1", faker.Sentence()))
	require.NoError(t, os.Setenv("test2", faker.Sentence()))
	current := NewCurrentEnvironment()
	envVars := current.GetEnvironmentVariables()
	assert.Len(t, envVars, 2)
	assert.False(t, envVars[0].Equal(envVars[1]))
}
