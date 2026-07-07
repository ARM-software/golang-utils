package retry

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/go-logr/logr"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

// RetryIf retries fn while retryConditionFn reports the returned error as
// retriable.
//
// The retry timing and attempt limits are controlled by retryPolicy. If the
// policy is disabled, fn is executed once without retrying.
func RetryIf(ctx context.Context, logger logr.Logger, retryPolicy *RetryPolicyConfiguration, fn func() error, msgOnRetry string, retryConditionFn func(err error) bool) error {
	if retryPolicy == nil {
		return commonerrors.New(commonerrors.ErrUndefined, "missing retry policy configuration")
	}
	if !retryPolicy.Enabled {
		return fn()
	}
	hasJitter := reflection.IsNotEmpty(retryPolicy.RetryMaxJitter)
	var retryType retry.DelayTypeFunc
	switch {
	case retryPolicy.LinearBackOffEnabled && hasJitter:
		retryType = retry.CombineDelay(retry.FixedDelay, retry.RandomDelay)
	case retryPolicy.LinearBackOffEnabled && !hasJitter:
		retryType = retry.FixedDelay
	case retryPolicy.BackOffEnabled && hasJitter:
		retryType = retry.CombineDelay(retry.BackOffDelay, retry.RandomDelay)
	case retryPolicy.BackOffEnabled && !hasJitter:
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
			retry.MaxJitter(retryPolicy.RetryMaxJitter),
			retry.DelayType(retryType),
			retry.Attempts(safecast.ToUint(retryPolicy.RetryMax)),
			retry.RetryIf(retryConditionFn),
			retry.LastErrorOnly(true),
			retry.Context(ctx),
		),
	)
}

// RetryOnError retries fn when the returned error matches any of retriableErr.
//
// The retry timing and attempt limits are controlled by retryPolicy.
func RetryOnError(ctx context.Context, logger logr.Logger, retryPolicy *RetryPolicyConfiguration, fn func() error, msgOnRetry string, retriableErr ...error) error {
	return RetryIf(ctx, logger, retryPolicy, fn, msgOnRetry, func(err error) bool {
		return commonerrors.Any(err, retriableErr...)
	})
}
