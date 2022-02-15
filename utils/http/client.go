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

func (c *GenericClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
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
