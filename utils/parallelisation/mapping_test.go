package parallelisation

import (
	"context"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestMapConcurrent(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx := context.Background()
	mapped, err := MapConcurrent(ctx, []int{1, 2, 3}, func(v int) string {
		return string(rune('0' + v))
	}, Workers(2))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"1", "2", "3"}, mapped)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = MapConcurrent(cancelledCtx, []int{1, 2, 3}, func(v int) int { return v })
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
}

func TestMapConcurrentRef(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx := context.Background()
	mapped, err := MapConcurrentRef(ctx, []int{1, 2, 3}, func(v *int) *string {
		if v == nil {
			return nil
		}
		result := string(rune('a' + *v - 1))
		return &result
	}, Workers(2))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, mapped)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = MapConcurrentRef(cancelledCtx, []int{1, 2, 3}, func(v *int) *int { return v })
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
}

func TestMapConcurrentSequence(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx := context.Background()
	sequence := iter.Seq[int](func(yield func(int) bool) {
		for _, value := range []int{1, 2, 3} {
			if !yield(value) {
				return
			}
		}
	})
	mapped, err := MapConcurrentSequence(ctx, sequence, func(v int) string {
		return string(rune('0' + v))
	}, Workers(2))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"1", "2", "3"}, mapped)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = MapConcurrentSequence(cancelledCtx, sequence, func(v int) int { return v })
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
}
