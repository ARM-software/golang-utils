package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/http/headers"
	"github.com/ARM-software/golang-utils/utils/http/httptest"
)

func TestClientWithHeadersWithDifferentBodies(t *testing.T) {
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
			client: func() IClient {
				return NewRetryableClient()
			},
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
		{
			clientName: "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == nil)",
			client: func() IClient {
				return NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), nil, logr.Discard(), "test-token")
			},
		},
		{
			clientName: "custom oauth client with retry after but no backoff using oauth2.Token (using custom client function with client == NewPlainHTTPClient())",
			client: func() IClient {
				return NewConfigurableRetryableOauthClientWithLoggerAndCustomClient(DefaultRobustHTTPClientConfigurationWithRetryAfter(), NewPlainHTTPClient().StandardClient(), logr.Discard(), "test-token")
			},
		},
		{
			clientName: "nil",
			client: func() IClient {
				return nil
			},
		},
	}

	tests := []struct {
		bodyType string
		uri      string
		headers  map[string]string
		body     interface{}
	}{
		{
			bodyType: "nil",
			uri:      "/foo/bar",
			headers:  nil,
			body:     nil,
		},
		{
			bodyType: "string",
			uri:      "/foo/bar",
			headers:  nil,
			body:     "some kind of string body",
		},
		{
			bodyType: "string reader",
			uri:      "/foo/bar",
			headers:  nil,
			body:     strings.NewReader("some kind of string body"),
		},
		{
			bodyType: "bytes",
			uri:      "/foo/bar",
			headers:  nil,
			body:     []byte("some kind of byte body"),
		},
		{
			bodyType: "byte buffer",
			uri:      "/foo/bar",
			headers:  nil,
			body:     bytes.NewBuffer([]byte("some kind of byte body")),
		},
		{
			bodyType: "byte reader",
			uri:      "/foo/bar",
			headers:  nil,
			body:     bytes.NewReader([]byte("some kind of byte body")),
		},
		{
			bodyType: "nil + single Host",
			uri:      "/foo/bar",
			headers: map[string]string{
				headers.HeaderHost: "example.com",
			},
			body: nil,
		},
		{
			bodyType: "string + WebSocket headers",
			uri:      "/foo/bar",
			headers: map[string]string{
				headers.HeaderConnection:          "Upgrade",
				headers.HeaderWebsocketVersion:    "13",
				headers.HeaderWebsocketKey:        "dGhlIHNhbXBsZSBub25jZQ==",
				headers.HeaderWebsocketProtocol:   "chat, superchat",
				headers.HeaderWebsocketExtensions: "permessage-deflate; client_max_window_bits",
			},
			body: "hello websocket",
		},
		{
			bodyType: "bytes + Sunset/Deprecation",
			uri:      "/foo/bar",
			headers: map[string]string{
				headers.HeaderSunset:      "2025-12-31T23:59:59Z",
				headers.HeaderDeprecation: "Tue, 01 Dec 2026 00:00:00 GMT",
			},
			body: []byte("payload with deprecation headers"),
		},
		{
			bodyType: "byte buffer + Link",
			uri:      "/foo/bar",
			headers: map[string]string{
				headers.HeaderLink: `<https://api.example.com/page2>; rel="next", <https://api.example.com/help>; rel="help"`,
			},
			body: bytes.NewBuffer([]byte("link header test")),
		},
		{
			bodyType: "reader + TUS upload headers",
			uri:      "/foo/bar",
			headers: map[string]string{
				headers.HeaderTusVersion:   "1.0.0",
				headers.HeaderUploadOffset: "1024",
				headers.HeaderUploadLength: "2048",
				headers.HeaderTusResumable: "1.0.0",
			},
			body: strings.NewReader("resumable upload content"),
		},
	}

	for i := range tests {
		test := tests[i]
		for j := range clientsToTest {
			defer goleak.VerifyNone(t)
			rawClient := clientsToTest[j]
			var headersSlice []string
			for k, v := range tests[i].headers {
				headersSlice = append(headersSlice, k, v)
			}
			client, err := NewHTTPClientWithUnderlyingClientWithHeaders(rawClient.client(), headersSlice...)
			require.NoError(t, err)
			defer func() { _ = client.Close() }()
			require.NotEmpty(t, client.StandardClient())
			t.Run(fmt.Sprintf("local host/Client %v/Body %v", rawClient.clientName, test.bodyType), func(t *testing.T) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				port := "28934"

				// Mock server which always responds 201.
				httptest.NewTestServer(t, ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}), port)
				time.Sleep(100 * time.Millisecond)
				url := fmt.Sprintf("http://127.0.0.1:%v/%v", port, test.uri)
				resp, err := client.Put(url, test.body)
				require.NoError(t, err)
				_ = resp.Body.Close()
				bodyReader, err := determineBodyReader(test.body)
				require.NoError(t, err)
				req, err := http.NewRequest("POST", url, bodyReader)
				require.NoError(t, err)
				resp, err = client.Do(req)
				require.NoError(t, err)
				_ = resp.Body.Close()
				cancel()
				time.Sleep(100 * time.Millisecond)
			})
			clientStruct, ok := client.(*ClientWithHeaders)
			require.True(t, ok)

			clientStruct.ClearHeaders()
			assert.Empty(t, clientStruct.headers)

			clientStruct.AppendHeader("hello", "world")
			require.NotEmpty(t, clientStruct.headers)
			header := clientStruct.headers.GetHeader("hello")
			require.NotNil(t, header)
			assert.Equal(t, headers.Header{Key: "hello", Value: "world"}, *header)

			clientStruct.RemoveHeader("hello")
			assert.Empty(t, clientStruct.headers)

			_ = client.Close()
		}
	}
}
