package parallelisation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/parallelisation/mocks"
)

func TestExecutionTimes(t *testing.T) {
	t.Run("clone", func(t *testing.T) {
		ctlr := gomock.NewController(t)
		defer ctlr.Finish()

		closerMock := mocks.NewMockCloser(ctlr)
		closerMock.EXPECT().Close().Return(nil).Times(6)
		group := NewCloserStoreWithOptions(ExecuteAll, Parallel, OnlyOnce, RetainAfterExecution)
		group.RegisterFunction(closerMock, closerMock, closerMock)
		c := group.Clone()
		require.NotNil(t, c)
		assert.Equal(t, group.Len(), c.Len())
		require.NoError(t, group.Close())
		closeClone, ok := c.(*CloserStore)
		require.True(t, ok)
		require.NoError(t, closeClone.Close())
	})

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

func TestCompoundGroup(t *testing.T) {
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	closerMock := mocks.NewMockCloser(ctlr)
	closerMock.EXPECT().Close().Return(nil).Times(17)
	group := NewCloserStoreWithOptions(ExecuteAll, OnlyOnce, Sequential)
	group.RegisterFunction(closerMock, closerMock, closerMock)

	compoundGroup := NewCompoundExecutionGroup(Parallel, RetainAfterExecution)
	compoundGroup.RegisterFunction(WrapCloseToContextualFunc(WrapCloserIntoCloseFunc(closerMock)))
	compoundGroup.RegisterExecutor(group)
	compoundGroup.RegisterFunction(WrapCancelToContextualFunc(WrapContextualToCancelFunc(WrapCloseToContextualFunc(WrapCloserIntoCloseFunc(closerMock)))))

	assert.Equal(t, 3, compoundGroup.Len())

	require.NoError(t, compoundGroup.Execute(context.Background()))
	require.NoError(t, compoundGroup.Execute(context.Background()))
	require.NoError(t, compoundGroup.Execute(context.Background()))
	require.NoError(t, compoundGroup.Execute(context.Background()))
	require.NoError(t, compoundGroup.Execute(context.Background()))
	require.NoError(t, compoundGroup.Execute(context.Background()))
	require.NoError(t, compoundGroup.Execute(context.Background()))

	t.Run("With cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, compoundGroup.Execute(ctx), commonerrors.ErrCancelled)
	})

}

func TestStoreOptions_MergeWithOptions(t *testing.T) {
	opts := WithOptions(Parallel).MergeWithOptions(OnlyOnce, ExecuteAll, Workers(5), Sequential)
	assert.True(t, opts.onlyOnce)
	assert.False(t, opts.stopOnFirstError)
	assert.True(t, opts.sequential)
	assert.Equal(t, 5, opts.workers)
}
