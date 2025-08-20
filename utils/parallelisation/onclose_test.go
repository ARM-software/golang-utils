package parallelisation

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/parallelisation/mocks"
)

//go:generate go tool mockgen -destination=./mocks/mock_$GOPACKAGE.go -package=mocks io.Closer
func TestCloseAll(t *testing.T) {
	t.Run("close", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).MinTimes(1)

		require.NoError(t, CloseAll(closerMock, closerMock, closerMock))
	})

	t.Run("close and join errors", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).MinTimes(1)

		require.NoError(t, CloseAllAndCollateErrors(closerMock, closerMock, closerMock))
	})

	t.Run("close with error", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()
		closeError := commonerrors.ErrUnexpected

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(closeError).MinTimes(1)

		errortest.AssertError(t, CloseAll(closerMock, closerMock, closerMock), closeError)
	})

	t.Run("close with errors", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()
		closeError := commonerrors.ErrUnexpected

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(closeError).MinTimes(1)

		errortest.AssertError(t, CloseAllAndCollateErrors(closerMock, closerMock, closerMock), closeError)
	})

	t.Run("close with 1 error", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected

		errortest.AssertError(t, CloseAllFunc(func() error { return nil }, func() error { return nil }, func() error { return closeError }, func() error { return nil }), closeError)
	})

	t.Run("close with 1 error but error collection", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected

		errortest.AssertError(t, CloseAllFuncAndCollateErrors(func() error { return nil }, func() error { return nil }, func() error { return closeError }, func() error { return nil }), closeError)
	})

}

func TestCloseOnce(t *testing.T) {
	t.Run("close every function once", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()
		closeError := commonerrors.ErrUnexpected

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(closeError).Times(3)

		group := NewCloseOnceGroup(Parallel, RetainAfterExecution)
		group.RegisterCloseFunction(WrapCloserIntoCloseFunc(closerMock), WrapCloserIntoCloseFunc(closerMock), WrapCloserIntoCloseFunc(closerMock))
		errortest.AssertError(t, group.Close(), closeError)
		require.NoError(t, group.Close())
	})
}

func TestCancelOnClose(t *testing.T) {
	t.Run("parallel", func(t *testing.T) {
		closeStore := NewCloseFunctionStoreStore(true)
		ctx1, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		ctx2, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		ctx3, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		assert.Equal(t, 3, closeStore.Len())
		require.NoError(t, DetermineContextError(ctx1))
		require.NoError(t, DetermineContextError(ctx2))
		require.NoError(t, DetermineContextError(ctx3))
		require.NoError(t, closeStore.Close())
		errortest.AssertError(t, DetermineContextError(ctx1), commonerrors.ErrCancelled)
		errortest.AssertError(t, DetermineContextError(ctx2), commonerrors.ErrCancelled)
		errortest.AssertError(t, DetermineContextError(ctx3), commonerrors.ErrCancelled)
	})
	t.Run("sequentially", func(t *testing.T) {
		closeStore := NewCloseFunctionStore(StopOnFirstError, Sequential)
		ctx1, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		ctx2, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		ctx3, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		assert.Equal(t, 3, closeStore.Len())
		require.NoError(t, DetermineContextError(ctx1))
		require.NoError(t, DetermineContextError(ctx2))
		require.NoError(t, DetermineContextError(ctx3))
		require.NoError(t, closeStore.Close())
		errortest.AssertError(t, DetermineContextError(ctx1), commonerrors.ErrCancelled)
		errortest.AssertError(t, DetermineContextError(ctx2), commonerrors.ErrCancelled)
		errortest.AssertError(t, DetermineContextError(ctx3), commonerrors.ErrCancelled)
	})
	t.Run("reverse", func(t *testing.T) {
		closeStore := NewCloseFunctionStore(StopOnFirstError, SequentialInReverse)
		ctx1, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		ctx2, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		ctx3, cancel := context.WithCancel(context.Background())
		closeStore.RegisterCancelFunction(cancel)
		assert.Equal(t, 3, closeStore.Len())
		require.NoError(t, DetermineContextError(ctx1))
		require.NoError(t, DetermineContextError(ctx2))
		require.NoError(t, DetermineContextError(ctx3))
		require.NoError(t, closeStore.Close())
		errortest.AssertError(t, DetermineContextError(ctx1), commonerrors.ErrCancelled)
		errortest.AssertError(t, DetermineContextError(ctx2), commonerrors.ErrCancelled)
		errortest.AssertError(t, DetermineContextError(ctx3), commonerrors.ErrCancelled)
	})
}

func TestSequentialExecution(t *testing.T) {
	tests := []struct {
		option StoreOption
	}{
		{StopOnFirstError},
		{JoinErrors},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("%v-%#v", i, test.option), func(t *testing.T) {
			opt := test.option(DefaultOptions())
			t.Run("sequentially", func(t *testing.T) {
				closeStore := NewCloseFunctionStore(test.option, Sequential)
				ctx1, cancel1 := context.WithCancel(context.Background())
				closeStore.RegisterCloseFunction(func() error { cancel1(); return DetermineContextError(ctx1) })
				ctx2, cancel2 := context.WithCancel(context.Background())
				closeStore.RegisterCloseFunction(func() error { cancel2(); return DetermineContextError(ctx2) })
				ctx3, cancel3 := context.WithCancel(context.Background())
				closeStore.RegisterCloseFunction(func() error { cancel3(); return DetermineContextError(ctx3) })
				assert.Equal(t, 3, closeStore.Len())
				require.NoError(t, DetermineContextError(ctx1))
				require.NoError(t, DetermineContextError(ctx2))
				require.NoError(t, DetermineContextError(ctx3))

				errortest.AssertError(t, closeStore.Close(), commonerrors.ErrCancelled)
				errortest.AssertError(t, DetermineContextError(ctx1), commonerrors.ErrCancelled)
				if opt.stopOnFirstError {
					assert.NoError(t, DetermineContextError(ctx2))
					assert.NoError(t, DetermineContextError(ctx3))
				} else {
					errortest.AssertError(t, DetermineContextError(ctx2), commonerrors.ErrCancelled)
					errortest.AssertError(t, DetermineContextError(ctx3), commonerrors.ErrCancelled)
				}

			})
			t.Run("reverse", func(t *testing.T) {
				closeStore := NewCloseFunctionStore(test.option, SequentialInReverse)
				ctx1, cancel1 := context.WithCancel(context.Background())
				closeStore.RegisterCloseFunction(func() error { cancel1(); return DetermineContextError(ctx1) })
				ctx2, cancel2 := context.WithCancel(context.Background())
				closeStore.RegisterCloseFunction(func() error { cancel2(); return DetermineContextError(ctx2) })
				ctx3, cancel3 := context.WithCancel(context.Background())
				closeStore.RegisterCloseFunction(func() error { cancel3(); return DetermineContextError(ctx3) })
				assert.Equal(t, 3, closeStore.Len())
				require.NoError(t, DetermineContextError(ctx1))
				require.NoError(t, DetermineContextError(ctx2))
				require.NoError(t, DetermineContextError(ctx3))
				errortest.AssertError(t, closeStore.Close(), commonerrors.ErrCancelled)
				if opt.stopOnFirstError {
					assert.NoError(t, DetermineContextError(ctx1))
					assert.NoError(t, DetermineContextError(ctx2))
				} else {
					errortest.AssertError(t, DetermineContextError(ctx1), commonerrors.ErrCancelled)
					errortest.AssertError(t, DetermineContextError(ctx2), commonerrors.ErrCancelled)
				}
				errortest.AssertError(t, DetermineContextError(ctx3), commonerrors.ErrCancelled)
			})
		})
	}
}
