package retry

// BackoffStrategy describes how retry delays should grow across attempts.
type BackoffStrategy int

//go:generate go tool enumer -type=BackoffStrategy -text -json -yaml

const (
	// NoBackoff disables backoff.
	NoBackoff BackoffStrategy = iota
	// FixedBackoff keeps the same delay for every retry.
	FixedBackoff
	// FixedBackoffOrRetryAfter keeps the same delay for every retry or considers the `Retry-After` header value if provided.
	FixedBackoffOrRetryAfter
	// LinearBackoff increases the delay linearly with the retry number.
	LinearBackoff
	// ExponentialBackoff doubles the delay after each retry until the max delay.
	ExponentialBackoff
)

// DefaultRetryPolicyConfiguration returns the package's default retry policy for
// the selected backoff strategy.
func DefaultRetryPolicyConfiguration(backoffStrategy BackoffStrategy) *RetryPolicyConfiguration {
	switch backoffStrategy {
	case NoBackoff:
		return DefaultNoRetryPolicyConfiguration()
	case FixedBackoff:
		return DefaultBasicRetryPolicyConfiguration()
	case FixedBackoffOrRetryAfter:
		return DefaultRobustRetryPolicyConfiguration()
	case LinearBackoff:
		return DefaultLinearBackoffRetryPolicyConfiguration()
	case ExponentialBackoff:
		return DefaultExponentialBackoffRetryPolicyConfiguration()
	default:
		return DefaultNoRetryPolicyConfiguration()
	}
}
