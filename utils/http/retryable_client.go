/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"net/http"
	"net/url"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
)

type RetryableClient struct {
	client *retryablehttp.Client
}

// NewRetryableClient creates a new http client which will retry failed requests with exponential backoff.
func NewRetryableClient() IClient {
	return &RetryableClient{client: retryablehttp.NewClient()}
}

// NewConfigurableRetryableClient creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff).
func NewConfigurableRetryableClient(cfg *HTTPClientConfiguration) IClient {
	subClient := &retryablehttp.Client{
		HTTPClient:   cleanhttp.DefaultPooledClient(),
		Logger:       nil,
		RetryWaitMin: cfg.RetryPolicy.RetryWaitMin,
		RetryWaitMax: cfg.RetryPolicy.RetryWaitMax,
		RetryMax:     cfg.RetryPolicy.RetryMax,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      BackOffPolicyFactory(&cfg.RetryPolicy),
	}
	if t, ok := subClient.HTTPClient.Transport.(*http.Transport); ok {
		setTransportConfiguration(cfg, t)
	}
	return &RetryableClient{client: subClient}
}

func (c *RetryableClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

func (c *RetryableClient) Post(url, bodyType string, body interface{}) (*http.Response, error) {
	return c.client.Post(url, bodyType, body)
}

func (c *RetryableClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.client.PostForm(url, data)
}

func (c *RetryableClient) StandardClient() *http.Client {
	return c.client.StandardClient()
}

func (c *RetryableClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	r, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}
	return c.client.Do(r)
}

func (c *RetryableClient) Delete(url string) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *RetryableClient) Put(url string, rawBody interface{}) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(http.MethodPut, url, rawBody)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *RetryableClient) Close() error {
	c.StandardClient().CloseIdleConnections()
	return nil
}
