/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/config"
)

type RetryPolicyConfiguration struct {
	// Enabled specifies whether this retry policy is enabled or not. If not, no retry will be performed.
	Enabled bool `mapstructure:"enabled"`
	// RetryMax represents the maximum number of retries
	RetryMax int `mapstructure:"max_retry"`
	// RetryAfterDisabled tells the client not to consider the `Retry-After` header (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After) returned by server.
	RetryAfterDisabled bool `mapstructure:"retry_after_disabled"`
	// RetryWaitMin specifies the minimum time to wait between retries.
	RetryWaitMin time.Duration
	// RetryWaitMax represents the maximum time to wait (only necessary if backoff is enabled).
	RetryWaitMax time.Duration
	// BackOffEnabled states whether backoff must be performed during retries (by default, exponential backoff is performed unless LinearBackoff is enabled).
	BackOffEnabled bool `mapstructure:"backoff_enabled"`
	// LinearBackOffEnabled forces to perform linear backoff instead of exponential backoff provided BackOffEnabled is set to true.
	LinearBackOffEnabled bool `mapstructure:"linear_backoff_enabled"`
}

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
		validation.Field(&cfg.RetryWaitMax, validation.Required.When(cfg.BackOffEnabled), validation.Min(time.Duration(0))),
		validation.Field(&cfg.BackOffEnabled, validation.Required.When(cfg.LinearBackOffEnabled)),
	)
}

// DefaultNoRetryPolicyConfiguration defines a configuration for no retry being performed.
func DefaultNoRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return &RetryPolicyConfiguration{
		Enabled:            false,
		RetryAfterDisabled: true,
	}
}

// DefaultBasicRetryPolicyConfiguration defines a configuration for basic retries i.e. retrying straight after a failure for maximum 4 attempts.
func DefaultBasicRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return &RetryPolicyConfiguration{
		Enabled:              true,
		RetryMax:             4,
		RetryAfterDisabled:   true,
		RetryWaitMin:         0,
		RetryWaitMax:         0,
		BackOffEnabled:       false,
		LinearBackOffEnabled: false,
	}
}

// DefaultRobustRetryPolicyConfiguration defines a configuration for basic retries but considering any `Retry-After` being returned by server.
func DefaultRobustRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return &RetryPolicyConfiguration{
		Enabled:              true,
		RetryMax:             4,
		RetryAfterDisabled:   false,
		RetryWaitMin:         0,
		RetryWaitMax:         0,
		BackOffEnabled:       false,
		LinearBackOffEnabled: false,
	}
}

// DefaultExponentialBackoffRetryPolicyConfiguration defines a configuration for retries with exponential backoff.
func DefaultExponentialBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return &RetryPolicyConfiguration{
		Enabled:              true,
		RetryMax:             4,
		RetryAfterDisabled:   false,
		RetryWaitMin:         time.Second,
		RetryWaitMax:         30 * time.Second,
		BackOffEnabled:       true,
		LinearBackOffEnabled: false,
	}
}

// DefaultLinearBackoffRetryPolicyConfiguration defines a configuration for retries with linear backoff.
func DefaultLinearBackoffRetryPolicyConfiguration() *RetryPolicyConfiguration {
	return &RetryPolicyConfiguration{
		Enabled:              true,
		RetryMax:             4,
		RetryAfterDisabled:   false,
		RetryWaitMin:         time.Second,
		RetryWaitMax:         time.Second, // See https://github.com/hashicorp/go-retryablehttp/blob/ff6d014e72d968e0f328637b209477ee09393175/client.go#L505
		BackOffEnabled:       true,
		LinearBackOffEnabled: true,
	}
}
