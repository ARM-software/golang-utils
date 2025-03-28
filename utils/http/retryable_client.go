/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// RetryableClient is an http client which will retry failed requests according to the retry configuration.
type RetryableClient struct {
	client *retryablehttp.Client
}

// NewRetryableClient creates a new http client which will retry failed requests with exponential backoff.
// It is based on `retryablehttp` default client
func NewRetryableClient() IRetryableClient {
	return &RetryableClient{client: retryablehttp.NewClient()}
}

// NewRetryableOauthClient creates a new http client with an authorisation token which will retry failed requests with exponential backoff.
func NewRetryableOauthClient(token string) IRetryableClient {
	return NewConfigurableRetryableOauthClient(DefaultRobustHTTPClientConfigurationWithExponentialBackOff(), token)
}

// NewRetryableOauthClientWithToken creates a new http client with an authorisation token which will retry failed requests with exponential backoff. It takes a full oauth2.Token to give more configuration for the token
func NewRetryableOauthClientWithToken(t *oauth2.Token) IRetryableClient {
	return NewConfigurableRetryableOauthClientWithToken(DefaultRobustHTTPClientConfigurationWithExponentialBackOff(), t)
}

// NewConfigurableRetryableOauthClientWithToken creates a new http client with an authorisation token which will retry based on the configuration.
// It takes a full oauth2.Token to give more configuration for the token
func NewConfigurableRetryableOauthClientWithToken(cfg *HTTPClientConfiguration, t *oauth2.Token) IRetryableClient {
	return NewConfigurableRetryableOauthClientWithTokenAndLogger(cfg, logr.Logger{}, t)
}

// NewConfigurableRetryableOauthClientWithTokenAndLogger creates a new http client with an authorisation token which will retry based on the configuration.
// It takes a logger to allow for debugging. It takes a full oauth2.Token to give more configuration for the token
func NewConfigurableRetryableOauthClientWithTokenAndLogger(cfg *HTTPClientConfiguration, logger logr.Logger, t *oauth2.Token) IRetryableClient {
	tc := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(t))
	return NewConfigurableRetryableClientWithLoggerFromClient(cfg, logger, tc)
}

// NewConfigurableRetryableOauthClient creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff) with the authorisation header set to token
func NewConfigurableRetryableOauthClient(cfg *HTTPClientConfiguration, token string) IRetryableClient {
	return NewConfigurableRetryableOauthClientWithLogger(cfg, logr.Logger{}, token)
}

// NewConfigurableRetryableOauthClientWithLogger creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff) with the authorisation header set to token
// It is also possible to supply a logger for debug purposes
func NewConfigurableRetryableOauthClientWithLogger(cfg *HTTPClientConfiguration, logger logr.Logger, token string) IRetryableClient {
	return NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(cfg, nil, logger, token)
}

// NewConfigurableRetryableOauthClientWithLoggerAndCustomClient creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff) with the authorisation header set to token
// The underlying client used by the oauth client can be optionally supplied via client
// It is also possible to supply a logger for debug purposes
func NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(cfg *HTTPClientConfiguration, client *http.Client, logger logr.Logger, token string) IRetryableClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	if client == nil {
		client = cleanhttp.DefaultPooledClient()
	}

	oauthClientCtx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
	tc := oauth2.NewClient(oauthClientCtx, ts)

	return NewConfigurableRetryableClientWithLoggerFromClient(cfg, logger, tc)
}

// NewClientWithAuthorisation creates a new http client with the authorisation header set according to the configuration
func NewClientWithAuthorisation(requestCfg *RequestConfiguration) IRetryableClient {
	return NewRetryableClientWithLogger(DefaultRobustHTTPClientConfigurationWithExponentialBackOff(), requestCfg, logr.Logger{})
}

// NewRetryableClientWithAuthorisation creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff) with the authorisation header set to token
func NewRetryableClientWithAuthorisation(cfg *HTTPClientConfiguration, requestCfg *RequestConfiguration) IRetryableClient {
	return NewRetryableClientWithLogger(cfg, requestCfg, logr.Logger{})
}

// NewRetryableClientWithLogger creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff) with the authorisation header set if enforced in configuration
// It is also possible to supply a logger for debug purposes
func NewRetryableClientWithLogger(cfg *HTTPClientConfiguration, requestCfg *RequestConfiguration, logger logr.Logger) IRetryableClient {
	return NewConfigurableRetryableClientWithLoggerAndCustomClient(cfg, requestCfg, logger, nil)
}

// NewConfigurableRetryableClientWithLoggerAndCustomClient creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff)
// with the authorisation header set to what the request configuration specifies.
// The underlying client used can be optionally supplied via client
// It is also possible to supply a logger for debug purposes
func NewConfigurableRetryableClientWithLoggerAndCustomClient(cfg *HTTPClientConfiguration, requestCfg *RequestConfiguration, logger logr.Logger, client *http.Client) IRetryableClient {
	if client == nil {
		client = cleanhttp.DefaultPooledClient()
	}

	tc := client
	if requestCfg != nil && requestCfg.Authorisation.Enforced {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: requestCfg.Authorisation.AccessToken,
				TokenType:   requestCfg.Authorisation.Scheme,
			},
		)
		oauthClientCtx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
		tc = oauth2.NewClient(oauthClientCtx, ts)
	}
	if requestCfg != nil && !reflection.IsEmpty(requestCfg.Retries) {
		cfg.RetryPolicy = requestCfg.Retries
	}

	return NewConfigurableRetryableClientWithLoggerFromClient(cfg, logger, tc)
}

// NewConfigurableRetryableClient creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff).
func NewConfigurableRetryableClient(cfg *HTTPClientConfiguration) IRetryableClient {
	return NewConfigurableRetryableClientWithLogger(cfg, logr.Logger{})
}

// NewConfigurableRetryableClientFromClient creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff).
// It also takes a custom client if you need an authenticated client
func NewConfigurableRetryableClientFromClient(cfg *HTTPClientConfiguration, client *http.Client) IRetryableClient {
	return NewConfigurableRetryableClientWithLoggerFromClient(cfg, logr.Logger{}, client)
}

// NewConfigurableRetryableClientWithLogger creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff).
// It is also possible to supply a logger for debug purposes
func NewConfigurableRetryableClientWithLogger(cfg *HTTPClientConfiguration, logger logr.Logger) IRetryableClient {
	return NewConfigurableRetryableClientWithLoggerFromClient(cfg, logger, cleanhttp.DefaultPooledClient())
}

// NewConfigurableRetryableClientWithLoggerFromClient creates a new http client which will retry failed requests according to the retry configuration (e.g. no retry, basic retry policy, exponential backoff).
// It is also possible to supply a logger for debug purposes as well as a custom client if you need an authenticated client
func NewConfigurableRetryableClientWithLoggerFromClient(cfg *HTTPClientConfiguration, logger logr.Logger, client *http.Client) IRetryableClient {
	subClient := &retryablehttp.Client{
		HTTPClient:      client,
		Logger:          newLogger(logger),
		RetryWaitMin:    cfg.RetryPolicy.RetryWaitMin,
		RetryWaitMax:    cfg.RetryPolicy.RetryWaitMax,
		RetryMax:        cfg.RetryPolicy.RetryMax,
		RequestLogHook:  nil,
		ResponseLogHook: nil,
		CheckRetry:      retryablehttp.DefaultRetryPolicy,
		Backoff:         BackOffPolicyFactory(&cfg.RetryPolicy).Apply,
		ErrorHandler:    nil,
	}
	if t, ok := subClient.HTTPClient.Transport.(*http.Transport); ok {
		setTransportConfiguration(cfg, t)
	}
	return &RetryableClient{client: subClient}
}

func (c *RetryableClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

func (c *RetryableClient) Options(url string) (*http.Response, error) {
	return c.doRetriableRequest(http.MethodOptions, url, nil)
}

func (c *RetryableClient) Post(url, contentType string, body interface{}) (*http.Response, error) {
	return c.client.Post(url, contentType, body)
}

func (c *RetryableClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.client.PostForm(url, data)
}

func (c *RetryableClient) StandardClient() *http.Client {
	return c.client.StandardClient()
}

func (c *RetryableClient) UnderlyingClient() *retryablehttp.Client {
	return c.client
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
	return c.doRetriableRequest(http.MethodDelete, url, nil)
}

func (c *RetryableClient) Put(url string, body interface{}) (*http.Response, error) {
	return c.doRetriableRequest(http.MethodPut, url, body)
}

func (c *RetryableClient) Close() error {
	c.StandardClient().CloseIdleConnections()
	return nil
}

func (c *RetryableClient) doRetriableRequest(method, url string, body interface{}) (*http.Response, error) {
	bodyReader, err := determineBodyReader(body)
	if err != nil {
		return nil, err
	}
	req, err := retryablehttp.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

type leveledLogger struct {
	logger logr.Logger
}

func (l *leveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(commonerrors.ErrUnexpected, msg, keysAndValues...)
}

func (l *leveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *leveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.V(1).Info(msg, keysAndValues...)
}

func (l *leveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.V(0).Info(fmt.Sprintf("WARNING: %v", msg), keysAndValues...)
}

func newLogger(logger logr.Logger) retryablehttp.LeveledLogger {
	if logger.IsZero() {
		return nil
	}
	return &leveledLogger{
		logger: logger,
	}
}
