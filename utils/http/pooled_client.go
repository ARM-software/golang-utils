/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

//  PooledClient is an HTTP client similar to
// http.Client, but with a shared Transport and different configuration values.
// It is based on https://github.com/hashicorp/go-cleanhttp which ensures the client configuration is only set for the current use case and not the whole project (i.e. no global variable)
type PooledClient struct {
	client *http.Client
}

//  NewDefaultPooledClient returns a new HTTP client with similar default values to
// http.Client, but with a shared Transport.
func NewDefaultPooledClient() IClient {
	return NewPooledClient(DefaultHTTPClientConfiguration())
}

//  NewFastPooledClient returns a new HTTP client with similar default values to
// fast http client https://github.com/valyala/fasthttp.
func NewFastPooledClient() IClient {
	return NewPooledClient(FastHTTPClientConfiguration())
}

//  NewPooledClient returns a new HTTP client using the configuration passed as argument.
// Do not use this function for
// transient clients as it can leak file descriptors over time. Only use this
// for clients that will be re-used for the same host(s).
func NewPooledClient(cfg *HTTPClientConfiguration) IClient {
	transport := cleanhttp.DefaultPooledTransport()
	setTransportConfiguration(cfg, transport)
	return &PooledClient{client: &http.Client{
		Transport: transport,
	}}
}

func (c *PooledClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

func (c *PooledClient) Post(url, bodyType string, body interface{}) (*http.Response, error) {
	if body == nil {
		return c.client.Post(url, bodyType, nil)
	}
	if b, ok := body.(io.Reader); ok {
		return c.client.Post(url, bodyType, b)
	}
	return nil, fmt.Errorf("%w: body is not an io.Reader", commonerrors.ErrInvalid)
}

func (c *PooledClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.client.PostForm(url, data)
}

func (c *PooledClient) StandardClient() *http.Client {
	return c.client
}

func (c *PooledClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

func (c *PooledClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *PooledClient) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *PooledClient) Put(url string, rawBody interface{}) (*http.Response, error) {
	var req *http.Request
	var err error
	if rawBody == nil {
		req, err = http.NewRequest("PUT", url, nil)
	} else {
		b, ok := rawBody.(io.Reader)
		if ok {
			req, err = http.NewRequest("PUT", url, b)
		} else {
			err = fmt.Errorf("%w: body is not an io.Reader", commonerrors.ErrInvalid)
		}
	}
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *PooledClient) Close() error {
	c.client.CloseIdleConnections()
	return nil
}
