package safeio

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func Test_convertIOError(t *testing.T) {
	assert.NoError(t, ConvertIOError(nil))
	err := errors.New("test")
	assert.ErrorIs(t, err, ConvertIOError(err))

	errortest.AssertError(t, ConvertIOError(commonerrors.ErrEOF), commonerrors.ErrEOF)
	errortest.AssertError(t, ConvertIOError(io.EOF), commonerrors.ErrEOF)
}
