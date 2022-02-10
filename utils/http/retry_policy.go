package http

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/hashicorp/go-retryablehttp"
)

// NoBackOff defines a no backoff retry policy.
// It is similar to a BasicRetry but also looks at the `Retry-After` header in the case of 429/503 HTTP errors (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After).
func NoBackOff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleep, found := findRetryAfter(resp)
	if found {
		return sleep
	}
	return BasicRetry(min, max, attemptNum, resp)
}

// BasicRetry defines a basic retry policy.
func BasicRetry(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	return min
}

// LinearBackoff defines a linear backoff retry policy based on the attempt number and with jitter to
// prevent a thundering herd.
// It is similar to retryablehttp.LinearJitterBackoff but it also looks at the `Retry-After` header in the case of 429/503 HTTP errors (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After).
func LinearBackoff(min, max time.Duration, attemptNum int, resp *http.Response, considerRetryAfter bool) time.Duration {
	if considerRetryAfter {
		sleep, found := findRetryAfter(resp)
		if found {
			return sleep
		}
	}
	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
}

// ExponentialBackoff defines an exponential backoff retry policy.
// It is exactly the same as retryablehttp.DefaultBackoff although the `Retry-After` header is checked differently to accept dates as well as time.
func ExponentialBackoff(min, max time.Duration, attemptNum int, resp *http.Response, considerRetryAfter bool) time.Duration {
	if considerRetryAfter {
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

// BackOffPolicyFactory generates a backoff policy based on configuration.
func BackOffPolicyFactory(cfg *RetryPolicyConfiguration) (policy retryablehttp.Backoff) {
	if cfg == nil || !cfg.Enabled || !cfg.BackOffEnabled {
		if cfg.RetryAfterDisabled {
			policy = BasicRetry
		} else {
			policy = NoBackOff
		}
		return
	}
	policy = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		if cfg.LinearBackOffEnabled {
			return LinearBackoff(min, max, attemptNum, resp, !cfg.RetryAfterDisabled)
		} else {
			return ExponentialBackoff(min, max, attemptNum, resp, !cfg.RetryAfterDisabled)
		}
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
						wait = afterTime.Sub(time.Now())
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
