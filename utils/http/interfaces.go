/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
// Package http provides a familiar HTTP client interface with
// various implementations in order to answer common use cases: e.g. need for a robust client (performs retries on server errors with different retry policies depending on configuration),
// for a fast client which leverages connection pools.
// It is a thin wrapper over some hashicorp implementations and hence over the standard net/http client library.
// This makes the client implementations very easy to drop into existing programs in lieu of standard library default client.
package http

import (
	"io"
	"net/http"
	"net/url"
)

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IClient

// IClient defines an HTTP client similar to http.Client but without shared state with other clients used in the same program.
// See https://github.com/hashicorp/go-cleanhttp for more details.
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
	// StandardClient returns a standard library *http.Client with a custom Transport layer.
	StandardClient() *http.Client
	// Put performs a PUT request.
	Put(url string, rawBody interface{}) (*http.Response, error)
	// Delete performs a DELETE request.
	Delete(url string) (*http.Response, error)
	// Do performs a generic request.
	Do(req *http.Request) (*http.Response, error)
}
