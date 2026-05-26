package filesystem

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestNewPathExistRule(t *testing.T) {
	t.Run("disable", func(t *testing.T) {
		err := NewOSPathExistRule(false).Validate(faker.URL())
		require.NoError(t, err)
	})
	t.Run("happy existing path", func(t *testing.T) {
		require.NoError(t, NewOSPathExistRule(true).Validate(TempDirectory()))
		testDir, err := TempDirInTempDir("test-path-rule-")
		require.NoError(t, err)
		defer func() { _ = Rm(testDir) }()
		require.NoError(t, NewOSPathExistRule(true).Validate(testDir))
		testFile, err := TouchTempFile(testDir, "test-file*.test")
		require.NoError(t, err)
		require.NoError(t, NewOSPathExistRule(true).Validate(testFile))
	})
	t.Run("non-existent path but valid", func(t *testing.T) {
		err := NewOSPathExistRule(true).Validate(strings.ReplaceAll(faker.Sentence(), " ", "/"))
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		err = NewOSPathValidationRule(true).Validate(strings.ReplaceAll(faker.Sentence(), " ", "/"))
		require.NoError(t, err)
		err = NewOSPathExistRule(true).Validate(faker.URL())
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		err = NewOSPathValidationRule(true).Validate(faker.URL())
		require.NoError(t, err)
	})

	t.Run("invalid paths", func(t *testing.T) {
		tests := []struct {
			entry         any
			expectedError []error
		}{
			{
				entry:         nil,
				expectedError: []error{commonerrors.ErrUndefined, commonerrors.ErrInvalid},
			},
			{
				entry:         "                  ",
				expectedError: []error{commonerrors.ErrUndefined, commonerrors.ErrInvalid},
			},
			{
				entry:         123,
				expectedError: []error{commonerrors.ErrInvalid},
			},
			{
				entry:         fmt.Sprintf("%v\n%v\n%v", faker.Paragraph(), faker.Paragraph(), faker.Sentence()),
				expectedError: []error{commonerrors.ErrInvalid},
			},
		}
		for i := range tests {
			test := tests[i]
			t.Run(fmt.Sprintf("%v", test.entry), func(t *testing.T) {
				err := NewOSPathValidationRule(true).Validate(test.entry)
				require.Error(t, err)
				errortest.AssertError(t, err, test.expectedError...)
				err = NewOSPathExistRule(true).Validate(test.entry)
				require.Error(t, err)
				errortest.AssertError(t, err, test.expectedError...)
			})
		}

	})
}

func TestNewPathExtensionRule(t *testing.T) {
	t.Run("disable", func(t *testing.T) {
		err := NewOSPathExtensionRule(false, ".json").Validate(faker.URL())
		require.NoError(t, err)
	})

	t.Run("happy path on global filesystem", func(t *testing.T) {
		require.NoError(t, NewOSPathExtensionRule(true, ".json", "yaml").Validate("config.json"))
		require.NoError(t, NewOSPathExtensionRule(true, ".json", "yaml").Validate("config.yaml"))
	})

	t.Run("happy path on custom filesystem", func(t *testing.T) {
		fs := NewTestFilesystem(t, '/')
		require.NoError(t, NewPathExtensionRule(fs, true, ".json", ".yaml").Validate("folder/config.yaml"))
	})

	t.Run("missing extension", func(t *testing.T) {
		err := NewOSPathExtensionRule(true, ".json").Validate("config")
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrNoExtension)
	})

	t.Run("wrong extension", func(t *testing.T) {
		err := NewOSPathExtensionRule(true, ".json").Validate("config.yaml")
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
	})

	t.Run("missing allowed extensions", func(t *testing.T) {
		err := NewOSPathExtensionRule(true).Validate("config.json")
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
	})
}
