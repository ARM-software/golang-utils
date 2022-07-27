/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/http/httptest"
)

// TestClientHappy tests that all requests are correctly populated and sent to the server. If requests are not as expected, an error is returned.
func TestClientHappy(t *testing.T) {
	clientsToTest := []struct {
		clientName string
		client     func() IClient
	}{
		{
			clientName: "default plain client",
			client:     NewPlainHTTPClient,
		},
		{
			clientName: "fast client",
			client:     NewFastPooledClient,
		},
		{
			clientName: "default pooled client",
			client:     NewDefaultPooledClient,
		},
		{
			clientName: "default retryable client",
			client:     NewRetryableClient,
		},
		{
			clientName: "client with no retry",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultHTTPClientConfiguration())
			},
		},
		{
			clientName: "client with basic retry",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultRobustHTTPClientConfiguration())
			},
		},
		{
			clientName: "client with exponential backoff",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithExponentialBackOff())
			},
		},
		{
			clientName: "client with linear backoff",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithLinearBackOff())
			},
		},
	}

	tests := []struct {
		method   string
		uri      string
		function func(client IClient, host string, uri string) (*http.Response, error)
	}{
		{
			method: http.MethodGet,
			uri:    "/foo/bar",
			function: func(client IClient, host string, uri string) (*http.Response, error) {
				return client.Get(fmt.Sprintf("%v/%v", host, uri))
			},
		},
		{
			method: http.MethodDelete,
			uri:    "/foo/bar",
			function: func(client IClient, host string, uri string) (*http.Response, error) {
				return client.Delete(fmt.Sprintf("%v/%v", host, uri))
			},
		},
		{
			method: http.MethodHead,
			uri:    "/foo/bar",
			function: func(client IClient, host string, uri string) (*http.Response, error) {
				return client.Head(fmt.Sprintf("%v/%v", host, uri))
			},
		},
		{
			method: http.MethodOptions,
			uri:    "/foo/bar",
			function: func(client IClient, host string, uri string) (*http.Response, error) {
				return client.Options(fmt.Sprintf("%v/%v", host, uri))
			},
		},
		{
			method: http.MethodPut,
			uri:    "/foo/bar",
			function: func(client IClient, host string, uri string) (*http.Response, error) {
				return client.Put(fmt.Sprintf("%v/%v", host, uri), nil)
			},
		},
		{
			method: http.MethodPost,
			uri:    "/foo/bar",
			function: func(client IClient, host string, uri string) (*http.Response, error) {
				return client.Post(fmt.Sprintf("%v/%v", host, uri), "", nil)
			},
		},
		{
			method: http.MethodPost,
			uri:    "/foo/bar",
			function: func(client IClient, host string, uri string) (*http.Response, error) {
				return client.PostForm(fmt.Sprintf("%v/%v", host, uri), nil)
			},
		},
	}

	for i := range tests {
		test := tests[i]
		for j := range clientsToTest {
			defer goleak.VerifyNone(t)
			client := clientsToTest[j]
			rawclient := client.client()
			defer func() { _ = rawclient.Close() }()
			t.Run(fmt.Sprintf("local host/Client %v/Method %v", client.clientName, test.method), func(t *testing.T) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Mock server which always responds 200, unless method is not as expected.
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					var statusCode int
					if r.Method == test.method && r.RequestURI == test.uri {
						statusCode = http.StatusOK
					} else {
						statusCode = http.StatusBadRequest
					}
					// The request succeeds
					w.WriteHeader(statusCode)
				})
				port := "28934"
				httptest.NewTestServer(t, ctx, handler, port)
				time.Sleep(100 * time.Millisecond)
				resp, err := test.function(rawclient, fmt.Sprintf("http://127.0.0.1:%v", port), test.uri)
				require.NoError(t, err)
				_ = resp.Body.Close()
				cancel()
				time.Sleep(100 * time.Millisecond)

			})
			_ = rawclient.Close()
		}
	}
}

// TestClientWithDifferentBodies tests that requests with different bodies are correctly populated.
func TestClientWithDifferentBodies(t *testing.T) {
	clientsToTest := []struct {
		clientName string
		client     func() IClient
	}{
		{
			clientName: "default plain client",
			client:     NewPlainHTTPClient,
		},
		{
			clientName: "fast client",
			client:     NewFastPooledClient,
		},
		{
			clientName: "default pooled client",
			client:     NewDefaultPooledClient,
		},
		{
			clientName: "default retryable client",
			client:     NewRetryableClient,
		},
		{
			clientName: "client with no retry",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultHTTPClientConfiguration())
			},
		},
		{
			clientName: "client with basic retry",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultRobustHTTPClientConfiguration())
			},
		},
		{
			clientName: "client with exponential backoff",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithExponentialBackOff())
			},
		},
		{
			clientName: "client with linear backoff",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithLinearBackOff())
			},
		},
	}

	tests := []struct {
		bodyType string
		uri      string
		body     interface{}
	}{
		{
			bodyType: "nil",
			uri:      "/foo/bar",
			body:     nil,
		},
		{
			bodyType: "string",
			uri:      "/foo/bar",
			body:     "some kind of string body",
		},
		{
			bodyType: "string reader",
			uri:      "/foo/bar",
			body:     strings.NewReader("some kind of string body"),
		},
		{
			bodyType: "bytes",
			uri:      "/foo/bar",
			body:     []byte("some kind of byte body"),
		},
		{
			bodyType: "byte buffer",
			uri:      "/foo/bar",
			body:     bytes.NewBuffer([]byte("some kind of byte body")),
		},
		{
			bodyType: "byte reader",
			uri:      "/foo/bar",
			body:     bytes.NewReader([]byte("some kind of byte body")),
		},
	}

	for i := range tests {
		test := tests[i]
		for j := range clientsToTest {
			defer goleak.VerifyNone(t)
			client := clientsToTest[j]
			rawclient := client.client()
			defer func() { _ = rawclient.Close() }()
			require.NotEmpty(t, rawclient.StandardClient())
			t.Run(fmt.Sprintf("local host/Client %v/Body %v", client.clientName, test.bodyType), func(t *testing.T) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				port := "28934"

				// Mock server which always responds 201.
				httptest.NewTestServer(t, ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}), port)
				time.Sleep(100 * time.Millisecond)
				url := fmt.Sprintf("http://127.0.0.1:%v/%v", port, test.uri)
				resp, err := rawclient.Put(url, test.body)
				require.NoError(t, err)
				_ = resp.Body.Close()
				bodyReader, err := determineBodyReader(test.body)
				require.NoError(t, err)
				req, err := http.NewRequest("POST", url, bodyReader)
				require.NoError(t, err)
				resp, err = rawclient.Do(req)
				require.NoError(t, err)
				_ = resp.Body.Close()
				cancel()
				time.Sleep(100 * time.Millisecond)

			})
			_ = rawclient.Close()
		}
	}
}
