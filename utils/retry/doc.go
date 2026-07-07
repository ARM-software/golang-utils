// Package retry provides configuration-driven retry helpers built on top of
// `github.com/avast/retry-go/v4`.
//
// It exposes a small retry policy model through [RetryPolicyConfiguration] and
// convenience helpers such as [RetryIf] and [RetryOnError]. It also provides a
// functional-option layer for composing retry policies in code.
package retry
