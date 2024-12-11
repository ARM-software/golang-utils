package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestSanetiseMajor(t *testing.T) {
	t.Run("No Version", func(t *testing.T) {
		v, err := SanetiseVersionMajor("")
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, v)
	})

	t.Run("Valid Version", func(t *testing.T) {
		v, err := SanetiseVersionMajor("1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1", v)
	})

	t.Run("Valid Version 2", func(t *testing.T) {
		v, err := SanetiseVersionMajor("v1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1", v)
	})

	t.Run("Invalid Version", func(t *testing.T) {
		v, err := SanetiseVersionMajor("aaaaaa")
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Empty(t, v)
	})
}

func TestSanetiseMajorMinor(t *testing.T) {
	t.Run("No Version", func(t *testing.T) {
		v, err := SanetiseVersionMajorMinor("")
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, v)
	})

	t.Run("Valid Version", func(t *testing.T) {
		v, err := SanetiseVersionMajorMinor("1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1.2", v)
	})

	t.Run("Valid Version 2", func(t *testing.T) {
		v, err := SanetiseVersionMajorMinor("v1.2.3")
		assert.NoError(t, err)
		assert.Equal(t, "1.2", v)
	})

	t.Run("Invalid Version", func(t *testing.T) {
		v, err := SanetiseVersionMajorMinor("aaaaaa")
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Empty(t, v)
	})
}
