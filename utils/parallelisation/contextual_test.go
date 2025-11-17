package parallelisation

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestForEach(t *testing.T) {
	cancelFunc := func() {}
	t.Run("close with 1 error", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected

		errortest.AssertError(t, ForEach(context.Background(), WithOptions(Parallel), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
	})

	t.Run("close with 1 error but error collection", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected
		errortest.AssertError(t, ForEach(context.Background(), WithOptions(Parallel), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
	})

	t.Run("close with 1 error and limited number of parallel workers", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected
		errortest.AssertError(t, ForEach(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
	})

	t.Run("close with 1 error but sequential", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected
		errortest.AssertError(t, ForEach(context.Background(), WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
		errortest.AssertError(t, BreakOnError(context.Background(), WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
		errortest.AssertError(t, BreakOnErrorOrEOF(context.Background(), WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
	})

	t.Run("close with cancellation", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, ForEach(cancelCtx, WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), commonerrors.ErrCancelled)
		errortest.AssertError(t, BreakOnError(cancelCtx, WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc)), commonerrors.ErrCancelled)
		errortest.AssertError(t, BreakOnErrorOrEOF(cancelCtx, WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc)), commonerrors.ErrCancelled)
	})

	t.Run("break on error with no error", func(t *testing.T) {
		require.NoError(t, BreakOnError(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc)))
	})
	t.Run("break on error or EOF with no error", func(t *testing.T) {
		require.NoError(t, BreakOnErrorOrEOF(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc)))
	})
	t.Run("break on error or EOF with no error", func(t *testing.T) {
		require.NoError(t, BreakOnErrorOrEOF(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), func(_ context.Context) error {
			return commonerrors.ErrEOF
		}))
		require.NoError(t, BreakOnErrorOrEOF(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), func(_ context.Context) error {
			return io.EOF
		}))
	})
	t.Run("for each with no error", func(t *testing.T) {
		require.NoError(t, ForEach(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc)))
	})
}

func TestDetermineContextError(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		require.NoError(t, DetermineContextError(context.Background()))
	})
	t.Run("cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, DetermineContextError(ctx))
		cancel()
		err := DetermineContextError(ctx)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})
	t.Run("cancellation with cause", func(t *testing.T) {
		cause := errors.New("a cause")
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(cause)
		require.NoError(t, DetermineContextError(ctx))
		cancel(cause)
		err := DetermineContextError(ctx)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		errortest.AssertErrorDescription(t, err, cause.Error())
	})
	t.Run("cancellation with timeout cause", func(t *testing.T) {
		cause := errors.New("a cause")
		ctx, cancel := context.WithTimeoutCause(context.Background(), 5*time.Second, cause)
		defer cancel()
		require.NoError(t, DetermineContextError(ctx))
		cancel()
		err := DetermineContextError(ctx)
		errortest.RequireError(t, err, commonerrors.ErrCancelled)
		assert.NotContains(t, err.Error(), cause.Error()) // the timeout did not take effect and a cancellation was performed instead so the cause is not passed through
	})
}
