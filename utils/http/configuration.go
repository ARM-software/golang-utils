/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-cleanhttp"

	"github.com/ARM-software/golang-utils/utils/config"
)

// HTTPClientConfiguration defines the client configuration. It can be used to tweak low level transport
// parameters in order to adapt the client to your use case.
// If unsure about the values to set, use the DefaultHTTPClientConfiguration or FastHTTPClientConfiguration depending on
// what flow you are dealing with.
type HTTPClientConfiguration struct {
	// MaxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialling,
	// active, and idle states. On limit violation, dials will block.
	//
	// Zero means no limit.
	MaxConnsPerHost int `mapstructure:"max_connections_per_host"`
	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns int `mapstructure:"max_idle_connections"`
	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `mapstructure:"max_idle_connections_per_host"`
	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	IdleConnTimeout time.Duration `mapstructure:"timeout_idle_connection"`
	// TLSHandshakeTimeout specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration `mapstructure:"timeout_tls_handshake"`
	// ExpectContinueTimeout, if non-zero, specifies the amount of
	// time to wait for a server's first response headers after fully
	// writing the request headers if the request has an
	// "Expect: 100-continue" header. Zero means no timeout and
	// causes the body to be sent immediately, without
	// waiting for the server to approve.
	// This time does not include the time to send the request header.
	ExpectContinueTimeout time.Duration `mapstructure:"timeout_expect_continue"`
	// RetryPolicy defines the retry policy to use for the retryable client
	RetryPolicy RetryPolicyConfiguration `mapstructure:"retry_policy"`
}

func (cfg *HTTPClientConfiguration) Validate() error {

	// Validate Embedded Structs
	err := config.ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.MaxIdleConns, validation.Min(0)),
		validation.Field(&cfg.MaxIdleConnsPerHost, validation.Max(cfg.MaxIdleConns)),
		validation.Field(&cfg.IdleConnTimeout, validation.Required),
		validation.Field(&cfg.RetryPolicy, validation.Required),
	)

}

// DefaultHTTPClientConfiguration uses default values similar to https://github.com/hashicorp/go-cleanhttp/blob/6d9e2ac5d828e5f8594b97f88c4bde14a67bb6d2/cleanhttp.go#L23
// Similar default values to http.DefaultTransport
func DefaultHTTPClientConfiguration() *HTTPClientConfiguration {
	defaultCfg := cleanhttp.DefaultPooledTransport()
	return &HTTPClientConfiguration{
		MaxConnsPerHost:       defaultCfg.MaxConnsPerHost,
		MaxIdleConns:          defaultCfg.MaxIdleConns,
		MaxIdleConnsPerHost:   defaultCfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       defaultCfg.IdleConnTimeout,
		TLSHandshakeTimeout:   defaultCfg.TLSHandshakeTimeout,
		ExpectContinueTimeout: defaultCfg.ExpectContinueTimeout,
		RetryPolicy:           *DefaultNoRetryPolicyConfiguration(),
	}
}

// FastHTTPClientConfiguration uses parameter values similar to https://github.com/valyala/fasthttp/blob/81fc96827033a5ee92d8a098ab1cdb9827e1eb8d/client.go
// the configuration was designed for some high performance edge cases: handle thousands of small to medium requests per seconds and a consistent low millisecond response time.
// It is, for example, used for checking authentication/authorisation.
func FastHTTPClientConfiguration() *HTTPClientConfiguration {
	return &HTTPClientConfiguration{
		MaxIdleConns:          1024,
		MaxIdleConnsPerHost:   512,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		RetryPolicy:           *DefaultNoRetryPolicyConfiguration(),
	}
}

// DefaultRobustHTTPClientConfiguration is similar to DefaultHTTPClientConfiguration but performs basic retry policy on failure.
func DefaultRobustHTTPClientConfiguration() *HTTPClientConfiguration {
	cfg := DefaultHTTPClientConfiguration()
	cfg.RetryPolicy = *DefaultBasicRetryPolicyConfiguration()
	return cfg
}

// DefaultRobustHTTPClientConfigurationWithRetryAfter is similar to DefaultRobustHTTPClientConfiguration but considers `Retry-After` header.
func DefaultRobustHTTPClientConfigurationWithRetryAfter() *HTTPClientConfiguration {
	cfg := DefaultHTTPClientConfiguration()
	cfg.RetryPolicy = *DefaultRobustRetryPolicyConfiguration()
	return cfg
}

// DefaultRobustHTTPClientConfigurationWithExponentialBackOff is similar to DefaultHTTPClientConfiguration but performs exponential backoff.
func DefaultRobustHTTPClientConfigurationWithExponentialBackOff() *HTTPClientConfiguration {
	cfg := DefaultHTTPClientConfiguration()
	cfg.RetryPolicy = *DefaultExponentialBackoffRetryPolicyConfiguration()
	return cfg
}

// DefaultRobustHTTPClientConfigurationWithLinearBackOff is similar to DefaultHTTPClientConfiguration but performs linear backoff.
func DefaultRobustHTTPClientConfigurationWithLinearBackOff() *HTTPClientConfiguration {
	cfg := DefaultHTTPClientConfiguration()
	cfg.RetryPolicy = *DefaultLinearBackoffRetryPolicyConfiguration()
	return cfg
}
