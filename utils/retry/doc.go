// Package retry provides configuration-driven retry helpers built on top of
// `github.com/avast/retry-go/v4`.
//
// It exposes a small retry policy model through [RetryPolicyConfiguration] and
// convenience helpers such as [RetryIf] and [RetryOnError]. It also provides a
// functional-option layer for composing retry policies in code.
//
// Backoff is the practice of waiting before the next retry attempt, often with
// the wait time increasing after repeated failures. Backoff is useful because
// it reduces pressure on a failing dependency, gives remote systems time to
// recover, and lowers the risk of many clients retrying at once after the same
// outage.
//
// References:
//   - Exponential backoff overview:
//     https://en.wikipedia.org/wiki/Exponential_backoff
//   - AWS Architecture Blog, Exponential Backoff and Jitter:
//     https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
//   - Google Cloud retry guidance:
//     https://cloud.google.com/storage/docs/retry-strategy
package retry
