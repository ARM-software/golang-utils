package http

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/hashicorp/go-retryablehttp"
)

// RetryWaitPolicy defines an `abstract` retry wait policy
type RetryWaitPolicy struct {
	ConsiderRetryAfter bool
}

// NewRetryWaitPolicy creates an generic RetryWaitPolicy based on configuration.
func NewRetryWaitPolicy(cfg *RetryPolicyConfiguration) *RetryWaitPolicy {
	if cfg == nil {
		return &RetryWaitPolicy{}
	}
	return &RetryWaitPolicy{
		ConsiderRetryAfter: !cfg.RetryAfterDisabled,
	}
}

// BasicRetryPolicy defines a basic retry policy i.e. it only waits a constant `min` amount of time between attempts.
// If enabled, it also looks at the `Retry-After` header in the case of 429/503 HTTP errors (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After).
type BasicRetryPolicy struct {
	RetryWaitPolicy
}

func (p *BasicRetryPolicy) Apply(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if p.ConsiderRetryAfter {
		sleep, found := findRetryAfter(resp)
		if found {
			return sleep
		}
	}
	return min
}

// NewBasicRetryPolicy creates a BasicRetryPolicy.
func NewBasicRetryPolicy(cfg *RetryPolicyConfiguration) IRetryWaitPolicy {
	return &BasicRetryPolicy{
		RetryWaitPolicy: *NewRetryWaitPolicy(cfg),
	}
}

// LinearBackoffPolicy defines a linear backoff retry policy based on the attempt number and with jitter to
// prevent a thundering herd.
// It is similar to retryablehttp.LinearJitterBackoff but if enabled, it also looks at the `Retry-After` header in the case of 429/503 HTTP errors (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After).
type LinearBackoffPolicy struct {
	RetryWaitPolicy
}

func (p *LinearBackoffPolicy) Apply(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if p.ConsiderRetryAfter {
		sleep, found := findRetryAfter(resp)
		if found {
			return sleep
		}
	}
	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
}

// NewLinearBackoffPolicy creates a LinearBackoffPolicy.
func NewLinearBackoffPolicy(cfg *RetryPolicyConfiguration) IRetryWaitPolicy {
	return &LinearBackoffPolicy{
		RetryWaitPolicy: *NewRetryWaitPolicy(cfg),
	}
}

// ExponentialBackoffPolicy defines an exponential backoff retry policy.
// It is exactly the same as retryablehttp.DefaultBackoff although the `Retry-After` header is checked differently to accept dates as well as time.
type ExponentialBackoffPolicy struct {
	RetryWaitPolicy
}

func (p *ExponentialBackoffPolicy) Apply(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if p.ConsiderRetryAfter {
		sleep, found := findRetryAfter(resp)
		if found {
			return sleep
		}
		return retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
	}
	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}
	return sleep
}

// NewExponentialBackoffPolicy creates a ExponentialBackoffPolicy.
func NewExponentialBackoffPolicy(cfg *RetryPolicyConfiguration) IRetryWaitPolicy {
	return &ExponentialBackoffPolicy{
		RetryWaitPolicy: *NewRetryWaitPolicy(cfg),
	}
}

// BackOffPolicyFactory generates a backoff policy based on configuration.
func BackOffPolicyFactory(cfg *RetryPolicyConfiguration) (policy IRetryWaitPolicy) {
	if cfg == nil || !cfg.Enabled || !cfg.BackOffEnabled {
		policy = NewBasicRetryPolicy(cfg)
		return
	}
	if cfg.LinearBackOffEnabled {
		policy = NewLinearBackoffPolicy(cfg)
	} else {
		policy = NewExponentialBackoffPolicy(cfg)
	}
	return
}

func findRetryAfter(resp *http.Response) (wait time.Duration, found bool) {
	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if s, ok := resp.Header[headers.RetryAfter]; ok {
				retryAfter := s[0]
				if sleep, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
					if sleep < 0 {
						sleep = 0
					}
					wait = time.Second * time.Duration(sleep)
					found = true
				}
				if afterTime, err := parseDate(retryAfter); err == nil {
					found = true
					if afterTime.After(time.Now()) {
						wait = time.Until(afterTime)
					} else {
						wait = time.Duration(0)
					}
				}
			}
		}
	}
	return
}

func parseDate(retryAfter string) (parsedTime time.Time, err error) {
	parsedTime, err = http.ParseTime(retryAfter)
	if err == nil {
		return
	}
	extraFormats := []string{time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano}
	for i := range extraFormats {
		parsedTime, err = time.Parse(extraFormats[i], retryAfter)
		if err == nil {
			return
		}
	}
	return
}
