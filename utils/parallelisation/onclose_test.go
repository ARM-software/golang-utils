package parallelisation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/parallelisation/mocks"
)

//go:generate go tool mockgen -destination=./mocks/mock_$GOPACKAGE.go -package=mocks io Closer
func TestCloseAll(t *testing.T) {
	t.Run("close", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).MinTimes(1)

		require.NoError(t, CloseAll(closerMock, closerMock, closerMock))
	})

	t.Run("close with error", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()
		closeError := commonerrors.ErrUnexpected

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(closeError).MinTimes(1)

		errortest.AssertError(t, CloseAll(closerMock, closerMock, closerMock), closeError)
	})

	t.Run("close with 1 error", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected

		errortest.AssertError(t, CloseAllFunc(func() error { return nil }, func() error { return nil }, func() error { return closeError }, func() error { return nil }), closeError)
	})

}
