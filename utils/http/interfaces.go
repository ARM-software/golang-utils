/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package http provides a familiar HTTP client interface with
// various implementations in order to answer common use cases: e.g. need for a robust client (performs retries on server errors with different retry policies depending on configuration),
// for a fast client which leverages connection pools.
// For the robust client, it is possible to set its configuration fully but this package also come with a set of preset configuration to answer most usecases
// e.g. to create a client which performs exponential backoff and listens to `Retry-After` headers, the following can be done:
// robustClient := NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithExponentialBackOff())
// resp, err := robustClient.Get("https://somehost.com")
// Whereas to create a client which will only perform 4 retries without backoff, the following can be done:
// retriableClient := NewConfigurableRetryableClient(DefaultRobustHTTPClientConfiguration())
// resp, err := retriableClient.Get("https://somehost.com")
// It is a thin wrapper over some hashicorp implementations and hence over the standard net/http client library.
// This makes the client implementations very easy to drop into existing programs in lieu of standard library default client.
package http

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IClient,IRetryWaitPolicy

// IClient defines an HTTP client similar to http.Client but without shared state with other clients used in the same program.
// See https://github.com/hashicorp/go-cleanhttp for more details.
type IClient interface {
	io.Closer
	// Get is a convenience helper for doing simple GET requests.
	Get(url string) (*http.Response, error)
	// Head is a convenience method for doing simple HEAD requests.
	Head(url string) (*http.Response, error)
	// Post is a convenience method for doing simple POST requests.
	Post(url, contentType string, body interface{}) (*http.Response, error)
	// PostForm is a convenience method for doing simple POST operations using
	// pre-filled url.Values form data.
	PostForm(url string, data url.Values) (*http.Response, error)
	// StandardClient returns a standard library *http.Client with a custom Transport layer.
	StandardClient() *http.Client
	// Put performs a PUT request.
	Put(url string, body interface{}) (*http.Response, error)
	// Delete performs a DELETE request.
	Delete(url string) (*http.Response, error)
	// Options performs an OPTIONS request.
	Options(url string) (*http.Response, error)
	// Do performs a generic request.
	Do(req *http.Request) (*http.Response, error)
}

// IRetryWaitPolicy defines the policy which specifies how much wait/sleep should happen between retry attempts.
type IRetryWaitPolicy interface {
	// Apply determines the amount of time to wait before the next retry attempt.
	// the time will be comprised between the `min` and `max` value unless other information are retrieved from the server response e.g. `Retry-After` header.
	Apply(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration
}

// IRetryableClient is a retryable client. It is a normal client with the additional method of extracting the underlying go-retryablehttp client so it can be used in libraries that use it
type IRetryableClient interface {
	IClient
	UnderlyingClient() *retryablehttp.Client
}
