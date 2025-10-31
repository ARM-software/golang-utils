/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package httptest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// NewTestServer creates a test server
func NewTestServer(t *testing.T, ctx context.Context, handler http.Handler, port string) {
	t.Helper()
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	require.NoError(t, err)
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: time.Minute,
		ReadTimeout:       time.Minute,
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	// Goroutine in charge of closing the server when requested.
	go func(ctx context.Context, server *http.Server) {
		<-ctx.Done()
		_ = server.Close()
	}(ctx, srv)

	go func(server *http.Server, listen net.Listener) {
		defer func() { _ = listen.Close() }()
		err = srv.Serve(listen)
	}(srv, list)
}
