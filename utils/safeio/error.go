package safeio

import (
	"io"
	"os"

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
	case commonerrors.Any(newErr, os.ErrClosed):
		// cancelling a reader on a copy will cause it to close the file and return os.ErrClosed so map it to cancelled for this package
		newErr = commonerrors.WrapError(commonerrors.ErrCancelled, newErr, "")
	}
	return
}
