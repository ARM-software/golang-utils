/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"io"
	"net/http"
	"net/url"
)

// IClient provides a familiar HTTP client interface with automatic retries and exponential backoff.
type IClient interface {
	io.Closer
	// Get is a convenience helper for doing simple GET requests.
	Get(url string) (*http.Response, error)
	// Head is a convenience method for doing simple HEAD requests.
	Head(url string) (*http.Response, error)
	// Post is a convenience method for doing simple POST requests.
	Post(url, bodyType string, body interface{}) (*http.Response, error)
	// PostForm is a convenience method for doing simple POST operations using
	// pre-filled url.Values form data.
	PostForm(url string, data url.Values) (*http.Response, error)
	// StandardClient returns a stdlib *http.Client with a custom Transport, which
	// shims in a *retryablehttp.Client for added retries.
	StandardClient() *http.Client
	// Perform a PUT request.
	Put(url string, rawBody interface{}) (*http.Response, error)
	// Perform a DELETE request.
	Delete(url string) (*http.Response, error)
	// Perform a generic request with exponential backoff
	Do(req *http.Request) (*http.Response, error)
}
