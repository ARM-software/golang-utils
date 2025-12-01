package environment

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/platform"
)

const (
	dotEnvPattern = ".env.*"
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
	t.Run("No dotenv files", func(t *testing.T) {
		os.Clearenv()
		value := faker.Sentence()
		require.NoError(t, os.Setenv("test1", value))
		require.NoError(t, os.Setenv("test2", faker.Sentence()))
		current := NewCurrentEnvironment()
		envVars := current.GetEnvironmentVariables()
		assert.Len(t, envVars, 2)
		assert.False(t, envVars[0].Equal(envVars[1]))
		found, err := FindEnvironmentVariableInEnvironment(current, nil, "test1")
		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.NotEmpty(t, found[0])
		assert.Equal(t, value, found[0].GetValue())
	})

	t.Run("With dotenv files", func(t *testing.T) {
		os.Clearenv()
		require.NoError(t, os.Setenv("test1", faker.Sentence()))
		require.NoError(t, os.Setenv("test2", faker.Sentence()))

		tmpDir, err := filesystem.TempDirInTempDir("dot-env")
		if err != nil {
			currentDir, err := os.Getwd()
			require.NoError(t, err)
			tmpDir, err = filesystem.TempDir(currentDir, "dot-env")
			require.NoError(t, err)
		}
		defer func() { _ = filesystem.Rm(tmpDir) }()
		dotenv1, err := filesystem.TempFile(tmpDir, dotEnvPattern)
		require.NoError(t, err)

		defer func() { _ = dotenv1.Close() }()
		test3 := NewEnvironmentVariable("test3", faker.Sentence())
		_, err = dotenv1.WriteString(test3.String())
		require.NoError(t, err)
		err = dotenv1.Close()
		require.NoError(t, err)

		dotenv2, err := filesystem.TempFile(tmpDir, dotEnvPattern)
		require.NoError(t, err)
		defer func() { _ = dotenv2.Close() }()
		test4 := NewEnvironmentVariable("test4", faker.Sentence())
		_, err = fmt.Fprintf(dotenv2, "%v\n", test4.String())
		require.NoError(t, err)
		test5 := NewEnvironmentVariable("test5", faker.Sentence())
		_, err = fmt.Fprintf(dotenv2, "%v\n", test5.String())
		require.NoError(t, err)
		err = dotenv2.Close()
		require.NoError(t, err)

		current := NewCurrentEnvironment()
		envVars := current.GetEnvironmentVariables(dotenv1.Name(), dotenv2.Name())
		SortEnvironmentVariables(envVars)
		assert.Len(t, envVars, 5)
		assert.False(t, envVars[0].Equal(envVars[1]))
		assert.True(t, envVars[2].Equal(test3))
		assert.True(t, envVars[3].Equal(test4))
		assert.True(t, envVars[4].Equal(test5))
	})
}

func Test_currentenv_GetEnvironmentVariable(t *testing.T) {
	t.Run("Env var exists", func(t *testing.T) {
		os.Clearenv()
		test := NewEnvironmentVariable(faker.Word(), faker.Sentence())
		require.NoError(t, os.Setenv(test.GetKey(), test.GetValue()))

		current := NewCurrentEnvironment()

		actual, err := current.GetEnvironmentVariable(test.GetKey())
		assert.NoError(t, err)
		assert.Equal(t, test, actual)
	})

	t.Run("Env var not exists", func(t *testing.T) {
		os.Clearenv()
		test := NewEnvironmentVariable(faker.Word(), faker.Sentence())
		current := NewCurrentEnvironment()

		actual, err := current.GetEnvironmentVariable(faker.Word())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.NotEqual(t, test, actual)
	})

	t.Run("With dotenv files", func(t *testing.T) {
		os.Clearenv()
		test1 := NewEnvironmentVariable("test1", faker.Sentence())
		test2 := NewEnvironmentVariable("test2", faker.Sentence())

		require.NoError(t, os.Setenv(test1.GetKey(), test1.GetValue()))
		require.NoError(t, os.Setenv(test2.GetKey(), test2.GetValue()))
		tmpDir, err := filesystem.TempDirInTempDir("dot-env")
		if err != nil {
			currentDir, err := os.Getwd()
			require.NoError(t, err)
			tmpDir, err = filesystem.TempDir(currentDir, "dot-env")
			require.NoError(t, err)
		}
		defer func() { _ = filesystem.Rm(tmpDir) }()
		dotenv1, err := filesystem.TempFile(tmpDir, dotEnvPattern)
		require.NoError(t, err)
		defer func() { _ = dotenv1.Close() }()
		test3 := NewEnvironmentVariable("test3", faker.Sentence())
		_, err = dotenv1.WriteString(test3.String())
		require.NoError(t, err)
		err = dotenv1.Close()
		require.NoError(t, err)

		dotenv2, err := filesystem.TempFile(tmpDir, dotEnvPattern)
		require.NoError(t, err)
		defer func() { _ = dotenv2.Close() }()
		test4 := NewEnvironmentVariable("test4", faker.Sentence())
		_, err = fmt.Fprintf(dotenv2, "%v\n", test4.String())
		require.NoError(t, err)
		test5 := NewEnvironmentVariable("test5", faker.Sentence())
		_, err = fmt.Fprintf(dotenv2, "%v\n", test5.String())
		require.NoError(t, err)
		var test6 IEnvironmentVariable
		if platform.IsWindows() {
			test6 = NewEnvironmentVariable("test6", "%test5%")
		} else {
			test6 = NewEnvironmentVariable("test6", "${test5}")
		}
		_, err = fmt.Fprintf(dotenv2, "%v\n", test6.String())
		require.NoError(t, err)
		err = dotenv2.Close()
		require.NoError(t, err)

		current := NewCurrentEnvironment()
		test1Actual, err := current.GetEnvironmentVariable(test1.GetKey(), dotenv1.Name(), dotenv2.Name())
		assert.NoError(t, err)
		assert.Equal(t, test1, test1Actual)
		test2Actual, err := current.GetEnvironmentVariable(test2.GetKey(), dotenv1.Name(), dotenv2.Name())
		assert.NoError(t, err)
		assert.Equal(t, test2, test2Actual)
		test3Actual, err := current.GetEnvironmentVariable(test3.GetKey(), dotenv1.Name(), dotenv2.Name())
		assert.NoError(t, err)
		assert.Equal(t, test3, test3Actual)
		test4Actual, err := current.GetEnvironmentVariable(test4.GetKey(), dotenv1.Name(), dotenv2.Name())
		assert.NoError(t, err)
		assert.Equal(t, test4, test4Actual)
		test5Actual, err := current.GetEnvironmentVariable(test5.GetKey(), dotenv1.Name(), dotenv2.Name())
		assert.NoError(t, err)
		assert.Equal(t, test5, test5Actual)
		test5Actual, err = current.GetExpandedEnvironmentVariable(test5.GetKey(), dotenv1.Name(), dotenv2.Name())
		assert.NoError(t, err)
		assert.Equal(t, test5, test5Actual)
		test6Actual, err := current.GetExpandedEnvironmentVariable(test6.GetKey(), dotenv1.Name(), dotenv2.Name())
		assert.NoError(t, err)
		assert.NotEqual(t, test6, test6Actual)
		assert.NotEqual(t, test5, test6Actual)
		assert.Equal(t, test5.GetValue(), test6Actual.GetValue())
		assert.NotEqual(t, test6.GetValue(), test6Actual.GetValue())

		os.Clearenv()

		test3Missing, err := current.GetEnvironmentVariable(test3.GetKey(), dotenv2.Name())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.NotEqual(t, test3, test3Missing)

		testMissing, err := current.GetEnvironmentVariable(faker.Word(), dotenv1.Name(), dotenv2.Name())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.Nil(t, testMissing)
	})
}

func Test_currentEnv_GetExpandedEnvironmentVariables(t *testing.T) {
	current := NewCurrentEnvironment()
	assert.NotEmpty(t, current.GetExpandedEnvironmentVariables())
}
