package safeio

import (
	"fmt"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func convertIOError(err error) error {
	err = commonerrors.ConvertContextError(err)
	if commonerrors.Any(err, io.EOF, io.ErrUnexpectedEOF, commonerrors.ErrEOF) {
		if err == commonerrors.ErrEOF {
			return err
		}
		return fmt.Errorf("%w: %v", commonerrors.ErrEOF, err.Error())
	}
	return err
}
