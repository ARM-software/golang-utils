/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package retry

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/config"
)

// RetryPolicyConfiguration configures retry attempts, wait timings, jitter, and
// backoff behaviour for the retry helpers in this package.
type RetryPolicyConfiguration struct {
	// Enabled specifies whether this retry policy is enabled. If false, no retry
	// is performed.
	Enabled bool `mapstructure:"enabled"`
	// RetryMax is the maximum number of attempts.
	RetryMax int `mapstructure:"max_retry"`
	// RetryAfterDisabled tells callers not to consider the `Retry-After` header (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After)
	// returned by a server.
	RetryAfterDisabled bool `mapstructure:"retry_after_disabled"`
	// RetryWaitMin specifies the minimum time to wait between retries.
	RetryWaitMin time.Duration `mapstructure:"min_wait"`
	// RetryWaitMax is the maximum time to wait between retries. It is primarily
	// used when backoff is enabled.
	RetryWaitMax time.Duration `mapstructure:"max_wait"`
	// RetryMaxJitter is the maximum random jitter added to retry delays.
	RetryMaxJitter time.Duration `mapstructure:"max_jitter"`
	// BackOffEnabled states whether backoff should be applied between retry
	// attempts (by default, exponential backoff is performed unless LinearBackoff is enabled).
	BackOffEnabled bool `mapstructure:"backoff_enabled"`
	// LinearBackOffEnabled switches from exponential to linear backoff when
	// BackOffEnabled is true.
	LinearBackOffEnabled bool `mapstructure:"linear_backoff_enabled"`
}

// Validate validates the retry policy configuration.
func (cfg *RetryPolicyConfiguration) Validate() error {

	// Validate Embedded Structs
	err := config.ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Enabled, validation.Required.When(cfg.BackOffEnabled || cfg.LinearBackOffEnabled)),
		validation.Field(&cfg.RetryMax, validation.Min(0), validation.Required.When(cfg.BackOffEnabled)),
		validation.Field(&cfg.RetryWaitMin, validation.Min(time.Duration(0))),
		validation.Field(&cfg.RetryMaxJitter, validation.Min(time.Duration(0))),
		validation.Field(&cfg.RetryWaitMax, validation.Required.When(cfg.BackOffEnabled), validation.Min(time.Duration(0))),
		validation.Field(&cfg.BackOffEnabled, validation.Required.When(cfg.LinearBackOffEnabled)),
	)
}

// DefaultNoRetryPolicyConfiguration returns a configuration with retries
// disabled.
func DefaultNoRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return &RetryPolicyConfiguration{
		Enabled:            false,
		RetryAfterDisabled: true,
		RetryMax:           0,
		RetryMaxJitter:     0,
	}
}

// DefaultBasicRetryPolicyConfiguration returns a basic retry configuration with
// four attempts and jitter but no backoff.
func DefaultBasicRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return WithOptions(WithRetryEnabled(), WithAttempts(4), WithJitterStrategy(25*time.Millisecond))(nil)
}

// DefaultRobustRetryPolicyConfiguration returns a basic retry configuration
// that also honours `Retry-After`.
func DefaultRobustRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return WithOptions(WithRetryAfterEnabled(), WithAttempts(4), WithJitterStrategy(25*time.Millisecond))(nil)
}

// DefaultFixedBackoffRetryPolicyConfiguration returns a retry configuration with
// four attempts and fixed backoff.
func DefaultFixedBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return WithOptions(WithRetryEnabled(), WithAttempts(4), WithFixedBackoff(time.Second))(nil)
}

// DefaultRobustFixedBackoffRetryPolicyConfiguration returns a retry
// configuration with four attempts and fixed backoff that also honours
// `Retry-After`.
func DefaultRobustFixedBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return WithOptions(WithRetryAfterEnabled(), WithAttempts(4), WithFixedBackoff(time.Second))(nil)
}

// DefaultExponentialBackoffRetryPolicyConfiguration returns a configuration for
// retries with exponential backoff.
func DefaultExponentialBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return WithOptions(WithRetryAfterEnabled(), WithAttempts(4), WithJitterStrategy(25*time.Millisecond), WithExponentialBackoff(time.Second, 30*time.Second))(nil)
}

// DefaultLinearBackoffRetryPolicyConfiguration returns a configuration for
// retries with linear backoff.
func DefaultLinearBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	// See https://github.com/hashicorp/go-retryablehttp/blob/ff6d014e72d968e0f328637b209477ee09393175/client.go#L505
	return WithOptions(WithRetryAfterEnabled(), WithAttempts(4), WithJitterStrategy(25*time.Millisecond), WithLinearBackoff(time.Second, time.Second))(nil)
}
