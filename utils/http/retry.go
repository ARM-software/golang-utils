package http

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/retry"
)

// RetryOnError allows the caller to retry fn when the error returned by fn is retriable
// as in of the type specified by retriableErr. backoff defines the maximum retries and the wait
// interval between two retries.
func RetryOnError(ctx context.Context, logger logr.Logger, retryPolicy *RetryPolicyConfiguration, fn func() error, msgOnRetry string, retriableErr ...error) error {
	return retry.RetryIf(ctx, logger, retryPolicy, fn, msgOnRetry, func(err error) bool {
		return commonerrors.Any(err, retriableErr...)
	})
}
