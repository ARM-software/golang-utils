package errortest

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestAssertError(t *testing.T) {
	AssertError(t, commonerrors.ErrUndefined, commonerrors.ErrNotFound, commonerrors.ErrMarshalling, commonerrors.ErrUndefined)
}

func TestRequireError(t *testing.T) {
	RequireError(t, commonerrors.ErrUndefined, commonerrors.ErrNotFound, commonerrors.ErrMarshalling, commonerrors.ErrUndefined)
}
