/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

// PooledClient is an HTTP client similar to
// http.Client, but with a shared Transport and different configuration values.
// It is based on https://github.com/hashicorp/go-cleanhttp which ensures the client configuration is only set for the current use case and not the whole project (i.e. no global variable)
type PooledClient struct {
	GenericClient
}

// NewDefaultPooledClient returns a new HTTP client with similar default values to
// http.Client, but with a shared Transport.
func NewDefaultPooledClient() IClient {
	return NewPooledClient(DefaultHTTPClientConfiguration())
}

// NewFastPooledClient returns a new HTTP client with similar default values to
// fast http client https://github.com/valyala/fasthttp.
func NewFastPooledClient() IClient {
	return NewPooledClient(FastHTTPClientConfiguration())
}

// NewPooledClient returns a new HTTP client using the configuration passed as argument.
// Do not use this function for
// transient clients as it can leak file descriptors over time. Only use this
// for clients that will be re-used for the same host(s).
func NewPooledClient(cfg *HTTPClientConfiguration) IClient {
	transport := cleanhttp.DefaultPooledTransport()
	setTransportConfiguration(cfg, transport)
	return NewGenericClient(&http.Client{
		Transport: transport,
	})
}
