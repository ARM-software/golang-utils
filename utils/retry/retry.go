package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-logr/logr"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

// RetryIf will retry fn when the value returned from retryConditionFn is true
func RetryIf(ctx context.Context, logger logr.Logger, retryPolicy *RetryPolicyConfiguration, fn func() error, msgOnRetry string, retryConditionFn func(err error) bool) error {
	if retryPolicy == nil {
		return fmt.Errorf("%w: missing retry policy configuration", commonerrors.ErrUndefined)
	}
	if !retryPolicy.Enabled {
		return fn()
	}
	var retryType retry.DelayTypeFunc
	switch {
	case retryPolicy.LinearBackOffEnabled:
		retryType = retry.CombineDelay(retry.FixedDelay, retry.RandomDelay)
	case retryPolicy.BackOffEnabled:
		retryType = retry.BackOffDelay
	default:
		retryType = retry.FixedDelay
	}

	return commonerrors.ConvertContextError(
		retry.Do(
			fn,
			retry.OnRetry(func(n uint, err error) {
				logger.Error(err, fmt.Sprintf("%v (attempt #%v)", msgOnRetry, n+1), "attempt", n+1)
			}),
			retry.Delay(retryPolicy.RetryWaitMin),
			retry.MaxDelay(retryPolicy.RetryWaitMax),
			retry.MaxJitter(25*time.Millisecond),
			retry.DelayType(retryType),
			retry.Attempts(safecast.ToUint(retryPolicy.RetryMax)),
			retry.RetryIf(retryConditionFn),
			retry.LastErrorOnly(true),
			retry.Context(ctx),
		),
	)
}

// RetryOnError allows the caller to retry fn when the error returned by fn is retriable
// as in of the type specified by retriableErr. backoff defines the maximum retries and the wait
// interval between two retries.
func RetryOnError(ctx context.Context, logger logr.Logger, retryPolicy *RetryPolicyConfiguration, fn func() error, msgOnRetry string, retriableErr ...error) error {
	return RetryIf(ctx, logger, retryPolicy, fn, msgOnRetry, func(err error) bool {
		return commonerrors.Any(err, retriableErr...)
	})
}
