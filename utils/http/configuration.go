/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"runtime"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/config"
)

type HTTPClientConfiguration struct {
	MaxConnsPerHost       int           `mapstructure:"max_connections_per_host"`
	MaxIdleConns          int           `mapstructure:"max_idle_connections"`
	MaxIdleConnsPerHost   int           `mapstructure:"max_idle_connections_per_host"`
	IdleConnTimeout       time.Duration `mapstructure:"timeout_idle_connection"`
	TLSHandshakeTimeout   time.Duration `mapstructure:"timeout_tls_handshake"`
	ExpectContinueTimeout time.Duration `mapstructure:"timeout_expect_continue"`
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
	)

}

// Default values similar to https://github.com/hashicorp/go-cleanhttp/blob/6d9e2ac5d828e5f8594b97f88c4bde14a67bb6d2/cleanhttp.go#L23
func DefaultHTTPClientConfiguration() *HTTPClientConfiguration {
	return &HTTPClientConfiguration{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// Default values similar to https://github.com/valyala/fasthttp/blob/81fc96827033a5ee92d8a098ab1cdb9827e1eb8d/client.go
func FastHTTPClientConfiguration() *HTTPClientConfiguration {
	return &HTTPClientConfiguration{
		MaxIdleConns:          1024,
		MaxIdleConnsPerHost:   512,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
