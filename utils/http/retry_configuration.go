/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"github.com/ARM-software/golang-utils/utils/retry"
)
// RetryPolicyConfiguration was moved to the `retry` module. Nonetheless, it was redefined here to avoid breaking changes
type RetryPolicyConfiguration = retry.RetryPolicyConfiguration

// DefaultNoRetryPolicyConfiguration defines a configuration for no retry being performed.
func DefaultNoRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return retry.DefaultNoRetryPolicyConfiguration()
}

// DefaultBasicRetryPolicyConfiguration defines a configuration for basic retries i.e. retrying straight after a failure for maximum 4 attempts.
func DefaultBasicRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return retry.DefaultBasicRetryPolicyConfiguration()
}

// DefaultRobustRetryPolicyConfiguration defines a configuration for basic retries but considering any `Retry-After` being returned by server.
func DefaultRobustRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return retry.DefaultRobustRetryPolicyConfiguration()
}

// DefaultExponentialBackoffRetryPolicyConfiguration defines a configuration for retries with exponential backoff.
func DefaultExponentialBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return retry.DefaultExponentialBackoffRetryPolicyConfiguration()
}

// DefaultLinearBackoffRetryPolicyConfiguration defines a configuration for retries with linear backoff.
func DefaultLinearBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return retry.DefaultLinearBackoffRetryPolicyConfiguration()
}
