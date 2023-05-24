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

// RequireError requires that the error is matching one of the `expectedErrors`
// This is a wrapper for commonerrors.Any.
func RequireError(t *testing.T, err error, expectedErrors ...error) {
	t.Helper()
	if commonerrors.Any(err, expectedErrors...) {
		return
	}
	t.FailNow()
}
