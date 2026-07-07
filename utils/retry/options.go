package retry

import (
	"time"

	"github.com/ARM-software/golang-utils/utils/collection"
)

// RetryOption configures a [RetryPolicyConfiguration].
type RetryOption func(*RetryPolicyConfiguration) *RetryPolicyConfiguration

// WithRetryEnabled turns retrying on for a policy that would otherwise execute
// only once.
func WithRetryEnabled() RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		if options == nil {
			options = DefaultNoRetryPolicyConfiguration()
		}
		options.Enabled = true
		return options
	}
}

// WithRetryAfterEnabled turns retrying on and allows callers to honour
// `Retry-After` response headers when they are available.
//
// This is useful for retry flows that interact with remote systems which may
// explicitly tell the client when to try again. `Retry-After` is an HTTP
// response header whose value indicates either how many seconds to wait before
// the next attempt or the time after which a retry is allowed, and it is
// commonly returned with HTTP 429 or 503 responses.
func WithRetryAfterEnabled() RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		cfg := WithRetryEnabled()(options)
		cfg.RetryAfterDisabled = false
		return cfg
	}
}

// WithAttempts limits how many total attempts are made, including the first.
//
// This is useful for bounding retry cost and ensuring failures surface after a
// predictable number of tries.
func WithAttempts(attempts int) RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		cfg := WithRetryEnabled()(options)
		cfg.RetryMax = attempts
		return cfg
	}
}

// WithFixedBackoff keeps the wait time the same between retry attempts.
//
// This is useful when a simple constant pause is sufficient and you do not want
// the delay to grow over time.
func WithFixedBackoff(delay time.Duration) RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		cfg := WithRetryEnabled()(options)
		cfg.RetryWaitMin = delay
		cfg.RetryWaitMax = delay
		return cfg
	}
}

// WithLinearBackoff increases the delay by a fixed step on each retry, capped
// at maxDelay.
//
// This is useful when retries should back off more gently than exponential
// backoff while still reducing pressure on the downstream system.
func WithLinearBackoff(delay, maxDelay time.Duration) RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		cfg := WithRetryEnabled()(options)
		cfg.RetryWaitMin = delay
		cfg.RetryWaitMax = maxDelay
		cfg.LinearBackOffEnabled = true
		cfg.BackOffEnabled = true
		return cfg
	}
}

// WithExponentialBackoff increases retry delays exponentially, capped at
// maxDelay.
//
// This is useful when repeated failures should quickly slow retry traffic to
// avoid overloading a struggling dependency.
func WithExponentialBackoff(delay, maxDelay time.Duration) RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		cfg := WithRetryEnabled()(options)
		cfg.RetryWaitMin = delay
		cfg.RetryWaitMax = maxDelay
		cfg.BackOffEnabled = true
		cfg.LinearBackOffEnabled = false
		return cfg
	}
}

// WithJitterStrategy adds a random component to retry delays.
//
// Jitter helps prevent many callers from retrying in lockstep, which reduces
// thundering-herd behaviour after a shared failure.
func WithJitterStrategy(maxJitter time.Duration) RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		cfg := WithRetryEnabled()(options)
		cfg.RetryWaitMax = maxJitter
		return cfg
	}
}

// WithRetryBudget sets a retry budget for the policy.
//
// A retry budget is a limit on how much time or capacity a call is allowed to
// spend on retries before giving up, so that retries do not grow without bound
// and crowd out useful work.
func WithRetryBudget(budget time.Duration) RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		cfg := WithRetryEnabled()(options)
		cfg.RetryWaitMax = budget
		return cfg
	}
}

// WithOptions combines multiple retry options into a reusable composite option
// applied in the order supplied.
//
// This is useful for defining named retry profiles that can be shared across
// callers without repeating the same option list.
func WithOptions(opts ...RetryOption) RetryOption {
	return func(options *RetryPolicyConfiguration) *RetryPolicyConfiguration {
		if options == nil {
			options = DefaultNoRetryPolicyConfiguration()
		}
		collection.ForEach(opts, func(opt RetryOption) {
			options = opt(options)
		})
		return options
	}
}
