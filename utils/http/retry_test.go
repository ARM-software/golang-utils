package http

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
)

func TestRetryOnError(t *testing.T) {
	tests := []struct {
		policy *RetryPolicyConfiguration
	}{
		{
			policy: DefaultBasicRetryPolicyConfiguration(),
		},
		{
			policy: DefaultNoRetryPolicyConfiguration(),
		},
		{
			policy: DefaultLinearBackoffRetryPolicyConfiguration(),
		},
		{
			policy: DefaultExponentialBackoffRetryPolicyConfiguration(),
		},
	}
	attemptCount := atomic.NewInt32(0)
	fn := func() error {
		attemptCount.Inc()
		return commonerrors.ErrUnknown
	}

	for i := range tests {
		test := tests[i]
		attemptCount.Store(0)
		t.Run(fmt.Sprintf("retry %T policy", test.policy), func(t *testing.T) {
			err := RetryOnError(context.Background(), logstest.NewTestLogger(t), test.policy, fn, "failed fn()", commonerrors.ErrUndefined, commonerrors.ErrUnknown)
			assert.Error(t, err)
			errortest.AssertError(t, err, commonerrors.ErrUnknown)
			if test.policy.RetryMax == 0 {
				assert.Equal(t, 1, int(attemptCount.Load()))
			} else {
				assert.Equal(t, test.policy.RetryMax, int(attemptCount.Load()))
			}
		})
		attemptCount.Store(0)
		t.Run(fmt.Sprintf("no retry when unlisted error (%T policy)", test.policy), func(t *testing.T) {
			err := RetryOnError(context.Background(), logstest.NewTestLogger(t), test.policy, fn, "failed fn()")
			assert.Error(t, err)
			errortest.AssertError(t, err, commonerrors.ErrUnknown)
			assert.Equal(t, 1, int(attemptCount.Load()))
		})
		attemptCount.Store(0)
		t.Run(fmt.Sprintf("no retry when unlisted error (%T policy)", test.policy), func(t *testing.T) {
			err := RetryOnError(context.Background(), logstest.NewTestLogger(t), test.policy, fn, "failed fn()", commonerrors.ErrMarshalling)
			assert.Error(t, err)
			errortest.AssertError(t, err, commonerrors.ErrUnknown)
			assert.Equal(t, 1, int(attemptCount.Load()))
		})
	}
	t.Run("check context cancellation", func(t *testing.T) {
		attemptCount.Store(0)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := RetryOnError(ctx, logstest.NewTestLogger(t), DefaultLinearBackoffRetryPolicyConfiguration(), fn, "failed fn()", commonerrors.ErrUndefined, commonerrors.ErrUnknown)
		assert.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.Zero(t, int(attemptCount.Load()))
	})
	t.Run("requires a retry policy", func(t *testing.T) {
		attemptCount.Store(0)
		err := RetryOnError(context.Background(), logstest.NewTestLogger(t), nil, fn, "failed fn()", commonerrors.ErrUnknown)
		assert.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
		assert.Zero(t, int(attemptCount.Load()))
	})
	t.Run("no retry if passing", func(t *testing.T) {
		attemptCount.Store(0)
		successFn := func() error {
			attemptCount.Inc()
			return nil
		}
		err := RetryOnError(context.Background(), logstest.NewTestLogger(t), DefaultLinearBackoffRetryPolicyConfiguration(), successFn, "failed fn()", commonerrors.ErrUndefined, commonerrors.ErrUnknown)
		require.NoError(t, err)
		assert.Equal(t, 1, int(attemptCount.Load()))
	})
}
