package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestSanitiseMajor(t *testing.T) {
	t.Run("No Version", func(t *testing.T) {
		v, err := SanitiseVersionMajor("")
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, v)
	})

	t.Run("Valid Version", func(t *testing.T) {
		v, err := SanitiseVersionMajor("1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1", v)
	})

	t.Run("Valid Version 2", func(t *testing.T) {
		v, err := SanitiseVersionMajor("v1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1", v)
	})

	t.Run("Invalid Version", func(t *testing.T) {
		v, err := SanitiseVersionMajor("aaaaaa")
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Empty(t, v)
	})
}

func TestSanitiseMajorMinor(t *testing.T) {
	t.Run("No Version", func(t *testing.T) {
		v, err := SanitiseVersionMajorMinor("")
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, v)
	})

	t.Run("Valid Version", func(t *testing.T) {
		v, err := SanitiseVersionMajorMinor("1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1.2", v)
	})

	t.Run("Valid Version 2", func(t *testing.T) {
		v, err := SanitiseVersionMajorMinor("v1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1.2", v)
	})

	t.Run("Invalid Version", func(t *testing.T) {
		v, err := SanitiseVersionMajorMinor("aaaaaa")
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Empty(t, v)
	})
}
