package semver

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/stretchr/testify/assert"
)

func TestCanonical(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Input    string
		Expected string
		Error    error
	}{
		{
			Name:     "full",
			Input:    "1.2.3",
			Expected: "1.2.3",
			Error:    nil,
		},
		{
			Name:     "full with build",
			Input:    "1.2.3+meta",
			Expected: "1.2.3",
			Error:    nil,
		},
		{
			Name:     "major minor",
			Input:    "1.2",
			Expected: "1.2.0",
			Error:    nil,
		},
		{
			Name:     "major",
			Input:    "1",
			Expected: "1.0.0",
			Error:    nil,
		},
		{
			Name:     "full with v",
			Input:    "v1.2.3",
			Expected: "1.2.3",
			Error:    nil,
		},
		{
			Name:     "full with v and build",
			Input:    "v1.2.3+meta",
			Expected: "1.2.3",
			Error:    nil,
		},
		{
			Name:     "major minor with v",
			Input:    "v1.2.0",
			Expected: "1.2.0",
			Error:    nil,
		},
		{
			Name:     "major with v",
			Input:    "v1",
			Expected: "1.0.0",
			Error:    nil,
		},
		{
			Name:     "empty",
			Input:    "",
			Expected: "",
			Error:    commonerrors.ErrUndefined,
		},
		{
			Name:     "invalid",
			Input:    "asdsdajk",
			Expected: "",
			Error:    commonerrors.ErrInvalid,
		},
		{
			Name:     "invalid 2",
			Input:    "1.2.3.4",
			Expected: "",
			Error:    commonerrors.ErrInvalid,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			expected, err := Canonical(test.Input)
			if test.Error == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, test.Error)
			}
			assert.Equal(t, test.Expected, expected)
		})
	}
}

func TestCanonicalPrefix(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Input    string
		Expected string
		Error    error
	}{
		{
			Name:     "full",
			Input:    "1.2.3",
			Expected: "v1.2.3",
			Error:    nil,
		},
		{
			Name:     "full with build",
			Input:    "1.2.3+meta",
			Expected: "v1.2.3",
			Error:    nil,
		},
		{
			Name:     "major minor",
			Input:    "1.2",
			Expected: "v1.2.0",
			Error:    nil,
		},
		{
			Name:     "major",
			Input:    "1",
			Expected: "v1.0.0",
			Error:    nil,
		},
		{
			Name:     "full with v",
			Input:    "v1.2.3",
			Expected: "v1.2.3",
			Error:    nil,
		},
		{
			Name:     "full with v and build",
			Input:    "v1.2.3+meta",
			Expected: "v1.2.3",
			Error:    nil,
		},
		{
			Name:     "major minor with v",
			Input:    "v1.2.0",
			Expected: "v1.2.0",
			Error:    nil,
		},
		{
			Name:     "major with v",
			Input:    "v1",
			Expected: "v1.0.0",
			Error:    nil,
		},
		{
			Name:     "empty",
			Input:    "",
			Expected: "",
			Error:    commonerrors.ErrUndefined,
		},
		{
			Name:     "invalid 1",
			Input:    "asdsdajk",
			Expected: "",
			Error:    commonerrors.ErrInvalid,
		},
		{
			Name:     "invalid 2",
			Input:    "v1.2.3.4",
			Expected: "",
			Error:    commonerrors.ErrInvalid,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			expected, err := CanonicalPrefix(test.Input)
			if test.Error == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, test.Error)
			}
			assert.Equal(t, test.Expected, expected)
		})
	}
}
