package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/safecast"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

func TestProxy(t *testing.T) {
	content := faker.Paragraph()
	path := faker.URL()
	password := faker.Password()
	tests := []struct {
		request *http.Request
	}{
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), io.NopCloser(strings.NewReader(content))),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), strings.NewReader(content)),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), io.NopCloser(bytes.NewReader([]byte(content)))),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), bytes.NewReader([]byte(content))),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), io.NopCloser(bytes.NewBuffer([]byte(content)))),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), bytes.NewBuffer([]byte(content))),
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req := test.request
			req.Header.Set(headers.AccessControlAllowOrigin, faker.Word())
			req.Header.Set(headers.XHTTPMethodOverride, http.MethodPut)
			req.Header.Set(headers.Authorization, password)
			assert.NotEqual(t, req.URL.String(), path)
			_, err := ProxyRequest(nil, http.MethodPost, "/")
			errortest.AssertError(t, err, commonerrors.ErrUndefined)
			preq, err := ProxyRequest(req, "     ", path)
			require.NoError(t, err)
			require.NotNil(t, preq)
			assert.Equal(t, path, preq.URL.String())
			assert.Equal(t, http.MethodGet, preq.Method)
			assert.NotEmpty(t, preq.Header.Get(headers.AccessControlAllowOrigin))
			assert.NotEmpty(t, preq.Header.Get(headers.Authorization))
			assert.NotZero(t, preq.ContentLength)
			resp := generateTestResponseBasedOnRequest(t, preq)
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()
			w := httptest.NewRecorder()
			require.NoError(t, ProxyResponse(context.Background(), resp, w))
			proxiedResp := w.Result()
			defer func() { _ = proxiedResp.Body.Close() }()
			assert.Empty(t, w.Header().Get(headers.AccessControlAllowOrigin))
			assert.Equal(t, http.MethodPut, w.Header().Get(headers.XHTTPMethodOverride))
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			responseContent, err := safeio.ReadAll(context.Background(), proxiedResp.Body)
			require.NoError(t, err)
			assert.Equal(t, content, string(responseContent))
		})
	}
}

func TestEmptyResponse(t *testing.T) {
	path := faker.URL()
	tests := []struct {
		request *http.Request
	}{
		{
			httptest.NewRequest(http.MethodGet, faker.URL(), nil),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), http.NoBody),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), io.NopCloser(http.NoBody)),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), bytes.NewReader(nil)),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), io.NopCloser(bytes.NewBuffer(nil))),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), strings.NewReader("")),
		},
		{
			request: httptest.NewRequest(http.MethodGet, faker.URL(), io.NopCloser(strings.NewReader(""))),
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req := test.request
			assert.NotEqual(t, req.URL.String(), path)
			preq, err := ProxyRequest(req, http.MethodPost, path)
			require.NoError(t, err)
			require.NotNil(t, preq)
			assert.Equal(t, path, preq.URL.String())
			assert.Equal(t, http.MethodPost, preq.Method)
			assert.Zero(t, preq.ContentLength)

			resp := generateTestResponseBasedOnRequest(t, preq)
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()
			w := httptest.NewRecorder()
			require.NoError(t, ProxyResponse(context.Background(), resp, w))
			require.NoError(t, err)
			returnedResp := w.Result()
			assert.LessOrEqual(t, returnedResp.ContentLength, safecast.ToInt64(0))
			assert.Equal(t, http.StatusOK, returnedResp.StatusCode)
		})
	}
}

func loopTestHandler(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	require.NotNil(t, r)
	require.NotNil(t, w)
	for k, v := range r.Header {
		for h := range v {
			w.Header().Add(k, v[h])
		}
	}
	written, err := safeio.CopyDataWithContext(r.Context(), r.Body, w)
	require.NoError(t, err)
	w.Header().Add(headers.ContentLength, strconv.FormatInt(written, 10))
	w.WriteHeader(http.StatusOK)
}

func generateTestResponseBasedOnRequest(t *testing.T, r *http.Request) *http.Response {
	t.Helper()
	require.NotNil(t, r)
	w := httptest.NewRecorder()
	loopTestHandler(t, w, r)
	return w.Result()
}
