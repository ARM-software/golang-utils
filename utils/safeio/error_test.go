package safeio

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func Test_convertIOError(t *testing.T) {
	assert.NoError(t, ConvertIOError(nil))
	err := errors.New("test")
	assert.ErrorIs(t, err, ConvertIOError(err))

	assert.True(t, commonerrors.Any(ConvertIOError(commonerrors.ErrEOF), commonerrors.ErrEOF))
	assert.True(t, commonerrors.Any(ConvertIOError(io.EOF), commonerrors.ErrEOF))
}
