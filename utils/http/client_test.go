/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/http/httptest"
)

func TestClient(t *testing.T) {
	clientsToTest := []struct {
		clientName string
		client     func() IClient
	}{
		{
			clientName: "fast client",
			client: func() IClient {
				return NewFastPooledClient()
			},
		},
		{
			clientName: "default client",
			client: func() IClient {
				return NewDefaultPooledClient()
			},
		},
		{
			clientName: "default retryable client",
			client: func() IClient {
				return NewRetryableClient()
			},
		},
		{
			clientName: "retryable client",
			client: func() IClient {
				return NewConfigurableRetryableClient(DefaultHTTPClientConfiguration())
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

				// Mock server which always responds 200.
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					var statusCode int
					if r.Method == test.method && r.RequestURI == test.uri {
						statusCode = 200
					} else {
						statusCode = 400
					}
					// The request succeeds
					w.WriteHeader(statusCode)
				})
				port := "28934"
				httptest.NewTestServer(t, ctx, handler, port)
				time.Sleep(100 * time.Millisecond)
				resp, err := test.function(rawclient, fmt.Sprintf("http://127.0.0.1:%v", port), test.uri)
				require.Nil(t, err)
				_ = resp.Body.Close()
				cancel()
				time.Sleep(100 * time.Millisecond)

			})
			_ = rawclient.Close()
		}
	}
}
