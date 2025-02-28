package filesystem

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestConvertFileSystemError(t *testing.T) {
	errortest.AssertError(t, ConvertFileSystemError(errors.New("write /tmp/dot-env932398628/.env.896898355: bad file descriptor")), commonerrors.ErrConflict, commonerrors.ErrCondition)
	errortest.AssertError(t, ConvertFileSystemError(fmt.Errorf("write %v: bad file descriptor", faker.Sentence())), commonerrors.ErrConflict, commonerrors.ErrCondition)
	errortest.AssertError(t, ConvertFileSystemError(nil), nil)
	errortest.AssertError(t, ConvertFileSystemError(context.Canceled), commonerrors.ErrCancelled)
}
