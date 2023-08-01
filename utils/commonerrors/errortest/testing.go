package errortest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// AssertError asserts that the error is matching one of the `expectedErrors`
// This is a wrapper for commonerrors.Any.
func AssertError(t *testing.T, err error, expectedErrors ...error) bool {
	if commonerrors.Any(err, expectedErrors...) {
		return true
	}
	return assert.Fail(t, fmt.Sprintf("Failed error assertion:\n actual: %v\n expected: %+v", err, expectedErrors))
}

// AssertErrorDescription asserts that the error description corresponds to one of the `expectedErrorDescriptions`
// This is a wrapper for commonerrors.CorrespondTo.
func AssertErrorDescription(t *testing.T, err error, expectedErrorDescriptions ...string) bool {
	if commonerrors.CorrespondTo(err, expectedErrorDescriptions...) {
		return true
	}
	return assert.Fail(t, fmt.Sprintf("Failed error description assertion:\n actual: %v\n expected: %+v", err, expectedErrorDescriptions))
}

// RequireError requires that the error is matching one of the `expectedErrors`
// This is a wrapper for commonerrors.Any.
func RequireError(t *testing.T, err error, expectedErrors ...error) {
	t.Helper()
	if commonerrors.Any(err, expectedErrors...) {
		return
	}
	t.FailNow()
}

// RequireErrorDescription requires that the error description corresponds to one of the `expectedErrorDescriptions`
// This is a wrapper for commonerrors.CorrespondTo.
func RequireErrorDescription(t *testing.T, err error, expectedErrorDescriptions ...string) {
	t.Helper()
	if commonerrors.CorrespondTo(err, expectedErrorDescriptions...) {
		return
	}
	t.FailNow()
}
