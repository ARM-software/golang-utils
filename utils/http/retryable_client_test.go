/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"context"
	"fmt"
	"github.com/go-faker/faker/v4"
	"net/http"
	"testing"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
	"golang.org/x/oauth2"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/http/httptest"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
)

// This test will sleep for 50ms after the request is made asynchronously via RetryableClient
// as to test that the retries/backoff are carried out.
func TestClient_Delete_Backoff(t *testing.T) {
	testLogger := logstest.NewTestLogger(t)
	tests := []struct {
		retryableClient IClient
		clientName      string
		token           *string
	}{
		{
			retryableClient: NewRetryableClient(),
			clientName:      "retryablehttp default client",
		},
		{
			retryableClient: NewConfigurableRetryableClientWithLogger(DefaultRobustHTTPClientConfiguration(), testLogger),
			clientName:      "client with retry but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClientWithLogger(DefaultRobustHTTPClientConfigurationWithRetryAfter(), testLogger),
			clientName:      "client with retry after but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClientWithLogger(DefaultRobustHTTPClientConfigurationWithExponentialBackOff(), testLogger),
			clientName:      "client with exponential backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClientWithLogger(DefaultRobustHTTPClientConfigurationWithLinearBackOff(), testLogger),
			clientName:      "client with linear backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClientWithLogger(func() *HTTPClientConfiguration {
				cfg := DefaultRobustHTTPClientConfigurationWithExponentialBackOff()
				cfg.RetryPolicy.RetryAfterDisabled = true
				return cfg
			}(), testLogger),
			clientName: "client with exponential backoff but no retry-after",
		},
		{
			retryableClient: NewConfigurableRetryableClientWithLogger(func() *HTTPClientConfiguration {
				cfg := DefaultRobustHTTPClientConfigurationWithLinearBackOff()
				cfg.RetryPolicy.RetryAfterDisabled = true
				return cfg
			}(), testLogger),
			clientName: "client with linear backoff but no retry-after",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewRetryableOauthClient("test-token"),
			clientName:      "default client with oath",
		},
		{
			token: field.ToOptionalString("test-token"),
			retryableClient: NewClientWithAuthorisation(&RequestConfiguration{Authorisation: Auth{
				Enforced:    true,
				Scheme:      AuthorisationSchemeBearer,
				AccessToken: "test-token",
			}}),
			clientName: "default client with Authorisation",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewConfigurableRetryableOauthClientWithLogger(DefaultRobustHTTPClientConfiguration(), testLogger, "test-token"),
			clientName:      "custom oauth client with retry but no backoff",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewConfigurableRetryableOauthClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), "test-token "),
			clientName:      "custom oauth client with retry after but no backoff",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewConfigurableRetryableOauthClientWithToken(DefaultRobustHTTPClientConfigurationWithRetryAfter(), &oauth2.Token{AccessToken: "test-token"}),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), nil, logr.Discard(), "test-token"),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == nil)",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), NewPlainHTTPClient().StandardClient(), logr.Discard(), "test-token"),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == NewPlainHTTPClient())",
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("with client %v", test.clientName), func(t *testing.T) {

			defer goleak.VerifyNone(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Collect errors for goroutines
			errs := make(chan error)
			counter := atomic.NewInt32(0)
			// Mock server which returns an error on first attempt.
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.token != nil {
					require.Equal(t, fmt.Sprintf("Bearer %v", *test.token), r.Header.Get(headers.Authorization))
				}
				require.Equal(t, r.Method, http.MethodDelete)
				require.Equal(t, r.RequestURI, "/foo/bar")
				if counter.Inc() == int32(1) {
					w.Header().Add(headers.RetryAfter, "1") // Retry after 1 second
					w.WriteHeader(http.StatusServiceUnavailable)
				} else {
					// The request succeeds
					w.WriteHeader(http.StatusOK)
				}
			})
			// Standing up a server which will fail first request to simulate a server error on the first request so that we can properly test the retry/backoff
			httptest.NewTestServer(t, ctx, handler, "28934")
			doneCh := make(chan struct{})

			// Make the request asynchronously and deliberately fail the first request,
			// exponential backoff should pass on subsequent attempts of the same request
			go func() {
				defer close(doneCh)
				resp, err := test.retryableClient.Delete("http://127.0.0.1:28934/foo/bar")
				if err != nil {
					errs <- err
				} else {
					_ = resp.Body.Close()
				}
			}()

			<-doneCh
			// close errs once all children finish.
			close(errs)
			cancel()
			for err := range errs {
				require.NoError(t, err)
			}
		})
	}
}

// This test simulates a 404, to which the backoff should not apply as it should only care about networking problems.
// If the client is found to be retrying, the test will fail.
func TestClient_Get_Fail_Timeout(t *testing.T) {
	tests := []struct {
		retryableClient IClient
		clientName      string
		token           *string
	}{
		{
			retryableClient: NewRetryableClient(),
			clientName:      "retryablehttp default client",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfiguration()),
			clientName:      "client with retry but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithRetryAfter()),
			clientName:      "client with retry after but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithExponentialBackOff()),
			clientName:      "client with exponential backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithLinearBackOff()),
			clientName:      "client with linear backoff",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewRetryableOauthClient("test-token"),
			clientName:      "custom oauth retryablehttp default client with oath",
		},
		{
			token: field.ToOptionalString("test-token"),
			retryableClient: NewClientWithAuthorisation(&RequestConfiguration{Authorisation: Auth{
				Enforced:    true,
				Scheme:      AuthorisationSchemeBearer,
				AccessToken: "test-token",
			}}),
			clientName: "client with Authorisation",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewConfigurableRetryableOauthClient(DefaultRobustHTTPClientConfiguration(), "test-token"),
			clientName:      "custom oauth client with retry but no backoff",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewConfigurableRetryableOauthClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), "test-token "),
			clientName:      "custom oauth client with retry after but no backoff",
		},
		{
			token:           field.ToOptionalString("test-token"),
			retryableClient: NewConfigurableRetryableOauthClientWithToken(DefaultRobustHTTPClientConfigurationWithRetryAfter(), &oauth2.Token{AccessToken: "test-token"}),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), nil, logr.Discard(), "test-token"),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == nil)",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), NewPlainHTTPClient().StandardClient(), logr.Discard(), "test-token"),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == NewPlainHTTPClient())",
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("with client %v", test.clientName), func(t *testing.T) {
			defer goleak.VerifyNone(t)
			// Collect errors for goroutines
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			errs := make(chan error)

			// Mock server which always responds 200.
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.token != nil {
					require.Equal(t, fmt.Sprintf("Bearer %v", *test.token), r.Header.Get(headers.Authorization))
				}
				require.Equal(t, r.Method, http.MethodGet)
				require.Equal(t, r.RequestURI, "/foo/bar")
				// The request fails
				w.WriteHeader(404)
			})

			httptest.NewTestServer(t, ctx, handler, "28935")

			// Send the asynchronous request
			var resp *http.Response
			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				var err error
				resp, err = test.retryableClient.Get("http://127.0.0.1:28935/foo/bar") //nolint:bodyclose // False Positive: Is closed below
				if err != nil {
					errs <- err
				} else {
					_ = resp.Body.Close()
				}
			}()
			select {
			case <-doneCh:
				// close errs once all children finish.
				close(errs)
				for err := range errs {
					require.Nil(t, err)
				}
			case <-time.After(1000 * time.Millisecond):
				t.Fatalf("Client should fail instantly with a 404 response, not retrying!")
			}
		})
	}
}

func TestUnderlyingClient(t *testing.T) {
	tests := []struct {
		retryableClient IRetryableClient
		clientName      string
	}{
		{
			retryableClient: NewRetryableClient(),
			clientName:      "retryablehttp default client",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfiguration()),
			clientName:      "client with retry but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithRetryAfter()),
			clientName:      "client with retry after but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithExponentialBackOff()),
			clientName:      "client with exponential backoff",
		},
		{
			retryableClient: NewConfigurableRetryableClient(DefaultRobustHTTPClientConfigurationWithLinearBackOff()),
			clientName:      "client with linear backoff",
		},
		{
			retryableClient: NewClientWithAuthorisation(&RequestConfiguration{
				Authorisation: Auth{
					Enforced:    true,
					Scheme:      AuthorisationSchemeBearer,
					AccessToken: faker.Password(),
				},
			}),
			clientName: "client with Authorisation",
		},
		{
			retryableClient: NewClientWithAuthorisation(&RequestConfiguration{}),
			clientName:      "client without Authorisation",
		},
		{
			retryableClient: NewClientWithAuthorisation(nil),
			clientName:      "default client",
		},
		{
			retryableClient: NewRetryableOauthClient("test-token"),
			clientName:      "custom oauth retryablehttp default client with oath",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClient(DefaultRobustHTTPClientConfiguration(), "test-token"),
			clientName:      "custom oauth client with retry but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), "test-token "),
			clientName:      "custom oauth client with retry after but no backoff",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClientWithToken(DefaultRobustHTTPClientConfigurationWithRetryAfter(), &oauth2.Token{AccessToken: "test-token"}),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), nil, logr.Discard(), "test-token"),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == nil)",
		},
		{
			retryableClient: NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), NewPlainHTTPClient().StandardClient(), logr.Discard(), "test-token"),
			clientName:      "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == NewPlainHTTPClient())",
		},
	}

	for i := range tests {
		t.Run(tests[i].clientName, func(t *testing.T) {
			client := tests[i].retryableClient
			rawClient, ok := client.(*RetryableClient)
			require.True(t, ok)
			assert.Equal(t, client.UnderlyingClient(), rawClient.UnderlyingClient())
		})
	}
}
