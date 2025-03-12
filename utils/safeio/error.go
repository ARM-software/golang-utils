package safeio

import (
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// ConvertIOError converts an I/O error into common errors.
func ConvertIOError(err error) (newErr error) {
	if err == nil {
		return
	}
	newErr = commonerrors.ConvertContextError(err)
	switch {
	case commonerrors.Any(newErr, commonerrors.ErrEOF):
	case commonerrors.Any(newErr, io.EOF, io.ErrUnexpectedEOF):
		newErr = commonerrors.WrapError(commonerrors.ErrEOF, newErr, "")
	}
	return
}
