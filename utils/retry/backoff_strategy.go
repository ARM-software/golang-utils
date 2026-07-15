package retry

// BackoffStrategy describes how retry delays should grow across attempts.
//
// A backoff strategy controls whether retries happen immediately or wait before
// trying again, and if they wait, how that delay changes over time.
type BackoffStrategy int

//go:generate go tool enumer -type=BackoffStrategy -text -json -yaml

const (
	// NoBackoff performs basic retries but applies no backoff strategy between
	// attempts.
	//
	// In practice this means retrying again immediately, subject only to the
	// overall retry attempt limit.
	NoBackoff BackoffStrategy = iota
	// NoBackoffButRetryAfter disables local backoff but still honours
	// `Retry-After` when a server provides it.
	NoBackoffButRetryAfter
	// FixedBackoff keeps the same delay for every retry and no jitter is applied.
	//
	// This is useful when a dependency benefits from a short, stable pause but
	// does not need progressively longer waits.
	FixedBackoff
	// FixedBackoffOrRetryAfter keeps the same delay for every retry or considers
	// the `Retry-After` header value if provided.
	FixedBackoffOrRetryAfter
	// LinearBackoff increases the delay linearly with the retry number.
	//
	// This is useful when retries should slow down steadily without escalating as
	// aggressively as exponential backoff.
	LinearBackoff
	// ExponentialBackoff doubles the delay after each retry until the max delay.
	//
	// This is useful when repeated failures should quickly reduce retry pressure
	// on a struggling dependency.
	//
	// Reference: https://en.wikipedia.org/wiki/Exponential_backoff
	ExponentialBackoff
)

// DefaultRetryPolicyConfiguration returns the package's default retry policy for
// the selected backoff strategy.
func DefaultRetryPolicyConfiguration(backoffStrategy BackoffStrategy) *RetryPolicyConfiguration {
	switch backoffStrategy {
	case NoBackoff:
		return DefaultBasicRetryPolicyConfiguration()
	case NoBackoffButRetryAfter:
		return DefaultRobustRetryPolicyConfiguration()
	case FixedBackoff:
		return DefaultFixedBackoffRetryPolicyConfiguration()
	case FixedBackoffOrRetryAfter:
		return DefaultRobustFixedBackoffRetryPolicyConfiguration()
	case LinearBackoff:
		return DefaultLinearBackoffRetryPolicyConfiguration()
	case ExponentialBackoff:
		return DefaultExponentialBackoffRetryPolicyConfiguration()
	default:
		return DefaultNoRetryPolicyConfiguration()
	}
}
