package parallelisation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ARM-software/golang-utils/utils/parallelisation/mocks"
)

func TestExecutionTimes(t *testing.T) {

	t.Run("close only Once Parallel with retention", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(3)

		group := NewCloserStoreWithOptions(ExecuteAll, Parallel, OnlyOnce, RetainAfterExecution)
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})
	t.Run("close only Once Sequential with retention", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(3)

		group := NewCloserStoreWithOptions(ExecuteAll, OnlyOnce, Sequential, RetainAfterExecution)
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})
	t.Run("close only Once Parallel without retention", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(3)

		group := NewCloserStoreWithOptions(ExecuteAll, Parallel, OnlyOnce, ClearAfterExecution)
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})
	t.Run("close only Once Sequential without retention", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(3)

		group := NewCloserStoreWithOptions(ExecuteAll, OnlyOnce, Sequential, ClearAfterExecution)
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})

	t.Run("close Multiple times Parallel", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(21)
		group := NewCloserStoreWithOptions(ExecuteAll, AnyTimes, Parallel, RetainAfterExecution, Workers(3))
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})
	t.Run("close Multiple times Sequential", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(21)
		group := NewCloserStoreWithOptions(ExecuteAll, AnyTimes, Sequential, RetainAfterExecution)
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})

	t.Run("close Multiple times Parallel without retention", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(3)
		group := NewCloserStoreWithOptions(ExecuteAll, AnyTimes, Parallel, ClearAfterExecution, Workers(3))
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})
	t.Run("close Multiple times Sequential without retention", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(3)
		group := NewCloserStoreWithOptions(ExecuteAll, AnyTimes, Sequential, ClearAfterExecution)
		group.RegisterFunction(closerMock, closerMock, closerMock)
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
		require.NoError(t, group.Close())
	})
}
