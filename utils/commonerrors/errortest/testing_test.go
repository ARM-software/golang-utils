package errortest

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestAssertError(t *testing.T) {
	AssertError(t, commonerrors.ErrUndefined, commonerrors.ErrNotFound, commonerrors.ErrMarshalling, commonerrors.ErrUndefined)
}

func TestAssertErrorDescription(t *testing.T) {
	AssertErrorDescription(t, commonerrors.ErrUndefined, "undefined", "not found")
}

func TestRequireError(t *testing.T) {
	RequireError(t, commonerrors.ErrUndefined, commonerrors.ErrNotFound, commonerrors.ErrMarshalling, commonerrors.ErrUndefined)
}

func TestRequireErrorContent(t *testing.T) {
	RequireErrorDescription(t, commonerrors.ErrUndefined, "undefined", "not found")
}
