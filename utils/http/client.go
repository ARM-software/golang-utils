/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// GenericClient is an HTTP client similar to
// http.Client (in fact, entirely based on it), but with extended capabilities (e.g. Options, Delete and Put methods are defined)
// to cover most methods defined in HTTP.
type GenericClient struct {
	client *http.Client
}

// NewGenericClient returns a new HTTP client (GenericClient) based on a standard library client implementation.
// If the raw client is not provided, it will default to the plain HTTP client (See NewPlainHTTPClient).
func NewGenericClient(rawClient *http.Client) IClient {
	c := rawClient
	if c == nil {
		c = &http.Client{
			Transport: cleanhttp.DefaultPooledTransport(),
		}
	}
	return &GenericClient{
		client: c,
	}
}

// NewPlainHTTPClient creates an HTTP client similar to http.DefaultClient but with extended methods and no shared state.
func NewPlainHTTPClient() IClient {
	return NewGenericClient(nil)
}

func (c *GenericClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

func (c *GenericClient) Post(url, contentType string, rawBody interface{}) (*http.Response, error) {
	b, err := determineBodyReader(rawBody)
	if err != nil {
		return nil, err
	}
	return c.client.Post(url, contentType, b)
}

func (c *GenericClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.client.PostForm(url, data)
}

func (c *GenericClient) StandardClient() *http.Client {
	return c.client
}

func (c *GenericClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// Do sends the provided HTTP request using the underlying http.Client
// and returns the resulting response.
//
// Validation
// The method performs minimal defensive checks before dispatching:
//   - req must not be nil or empty
//   - req.URL must not be empty
//
// If either validation fails, an UndefinedVariable error is returned.
//
// Security considerations
// This method intentionally delegates full request validation and
// authorisation decisions to the caller. In particular, it does not
// enforce host allow-listing, scheme restrictions, redirect policies,
// or other SSRF mitigations.
//
// Static analysis tools (e.g. gosec G704) flag direct execution of
// externally influenced requests as a potential Server-Side Request
// Forgery (SSRF) risk. This behaviour is deliberate: callers are
// responsible for validating and constraining req before invoking Do.
func (c *GenericClient) Do(req *http.Request) (resp *http.Response, err error) {
	if reflection.IsEmpty(req) {
		err = commonerrors.UndefinedVariable("request")
		return
	}
	if reflection.IsEmpty(req.URL) || reflection.IsEmpty(req.URL.String()) {
		err = commonerrors.UndefinedVariable("request's target (URL)")
		return
	}
	resp, err = c.client.Do(req) //nolint:gosec //G704: SSRF via taint analysis (gosec) It is user's responsibility to check the request
	return
}

func (c *GenericClient) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *GenericClient) Put(url string, rawBody interface{}) (*http.Response, error) {
	b, err := determineBodyReader(rawBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, url, b)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *GenericClient) Options(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodOptions, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *GenericClient) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

func determineBodyReader(rawBody interface{}) (reader io.Reader, err error) {
	switch body := rawBody.(type) {
	case []byte:
		reader = bytes.NewReader(body)
	case *bytes.Buffer:
		reader = bytes.NewReader(body.Bytes())
	case string:
		reader = strings.NewReader(body)
	case *bytes.Reader:
	case io.ReadSeeker:
	case io.Reader:
		reader = body
	case nil:

	default:
		err = fmt.Errorf("%w: cannot handle body's type %T", commonerrors.ErrInvalid, rawBody)
	}
	return
}
