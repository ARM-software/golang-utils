/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/http/httptest"
)

// This test will sleep for 50ms after the request is made asynchronously via RetryableClient
// as to test that the exponential backoff is working.
func TestClient_Delete_Backoff(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Collect errors for goroutines
	errs := make(chan error)
	// Mock server which always responds 200.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, r.Method, http.MethodDelete)
		require.Equal(t, r.RequestURI, "/foo/bar")
		// The request succeeds
		w.WriteHeader(200)
	})

	doneCh := make(chan struct{})
	// Make the request asynchronously and deliberately fail the first request,
	// exponential backoff should pass on subsequent attempts of the same request
	go func() {
		defer close(doneCh)
		resp, err := NewRetryableClient().Delete("http://127.0.0.1:28934/foo/bar")
		if err != nil {
			errs <- err
		} else {
			_ = resp.Body.Close()
		}
	}()

	// We set this to simulate a first error on the first request so that we can properly test the backoff
	time.Sleep(50 * time.Millisecond)
	httptest.NewTestServer(t, ctx, handler, "28934")

	<-doneCh
	// close errs once all children finish.
	close(errs)
	cancel()
	for err := range errs {
		require.Nil(t, err)
	}
}

// This test simulates a 404, to which the backoff should not apply as it should only care about networking problems.
// If the client is found to be retrying, the test will fail.
func TestClient_Get_Fail_Timeout(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Collect errors for goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error)

	// Mock server which always responds 200.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		resp, err = NewRetryableClient().Get("http://127.0.0.1:28935/foo/bar") //nolint:bodyclose // False Positive: Is closed below
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
}
