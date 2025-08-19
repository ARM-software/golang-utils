package parallelisation

import (
	"context"
	"testing"

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
		errortest.AssertError(t, ForEach(context.Background(), WithOptions(Parallel, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
	})

	t.Run("close with 1 error but error collection", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected
		errortest.AssertError(t, ForEach(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
	})

	t.Run("close with 1 error but sequential", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected
		errortest.AssertError(t, ForEach(context.Background(), WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
		errortest.AssertError(t, BreakOnError(context.Background(), WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), closeError)
	})

	t.Run("close with cancellation", func(t *testing.T) {
		closeError := commonerrors.ErrUnexpected
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, ForEach(cancelCtx, WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCloseToContextualFunc(func() error { return closeError }), WrapCancelToContextualFunc(cancelFunc)), commonerrors.ErrCancelled)
		errortest.AssertError(t, BreakOnError(cancelCtx, WithOptions(SequentialInReverse, JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc)), commonerrors.ErrCancelled)
	})

	t.Run("break on error with no error", func(t *testing.T) {
		require.NoError(t, BreakOnError(context.Background(), WithOptions(Workers(5), JoinErrors), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc), WrapCancelToContextualFunc(cancelFunc)))
	})
}
