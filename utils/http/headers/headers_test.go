package headers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/http/schemes"
)

func TestParseAuthorizationHeader(t *testing.T) {
	random, err := faker.RandomInt(0, len(schemes.HTTPAuthorisationSchemes)-1, 1)
	require.NoError(t, err)
	fakescheme := schemes.HTTPAuthorisationSchemes[random[0]]
	r, err := http.NewRequest(http.MethodGet, faker.URL(), nil)
	require.NoError(t, err)
	var scheme, token string
	t.Run("empty authorization header", func(t *testing.T) {
		scheme, token, err := ParseAuthorizationHeader(r)
		require.Error(t, err)
		errortest.RequireError(t, err, commonerrors.ErrUndefined)
		assert.Empty(t, scheme)
		assert.Empty(t, token)
	})
	t.Run("invalid authorization header", func(t *testing.T) {
		require.NoError(t, SetAuthorisation(r, faker.Word()))
		scheme, token, err = ParseAuthorizationHeader(r)
		require.Error(t, err)
		assert.True(t, commonerrors.Any(err, commonerrors.ErrInvalid))
		assert.Empty(t, scheme)
		assert.Empty(t, token)
		require.NoError(t, SetAuthorisation(r, faker.Sentence()))
		scheme, token, err = ParseAuthorizationHeader(r)
		require.Error(t, err)
		errortest.RequireError(t, err, commonerrors.ErrInvalid)
		assert.Empty(t, scheme)
		assert.Empty(t, token)
	})
	faketoken := faker.Password()
	t.Run("valid authorization header", func(t *testing.T) {
		require.NoError(t, SetAuthorisationToken(r, fakescheme, faketoken))
		scheme, token, err = ParseAuthorizationHeader(r)
		require.NoError(t, err)
		assert.Equal(t, fakescheme, scheme)
		assert.Equal(t, faketoken, token)
	})
	t.Run("valid authorisation header for websocket (workaround1)", func(t *testing.T) {
		value, err := GenerateAuthorizationHeaderValue(fakescheme, faketoken)
		require.NoError(t, err)
		tests := []struct {
			encoded string
		}{
			{
				encoded: base64.StdEncoding.EncodeToString([]byte(value)),
			},
			{
				encoded: base64.URLEncoding.EncodeToString([]byte(value)),
			},
			{
				encoded: base64.RawStdEncoding.EncodeToString([]byte(value)),
			},
			{
				encoded: base64.RawURLEncoding.EncodeToString([]byte(value)),
			},
		}
		for i := range tests {
			test := tests[i]
			t.Run("base64 encoding", func(t *testing.T) {
				r, err = http.NewRequest(http.MethodGet, faker.URL(), nil)
				require.NoError(t, err)
				r.Header.Add(HeaderWebsocketProtocol, "base64.binary.k8s.io")
				r.Header.Add(HeaderWebsocketProtocol, fmt.Sprintf("base64url.bearer.authorization.k8s.io.%v", test.encoded))
				scheme, token, err = ParseAuthorizationHeader(r)
				require.NoError(t, err)
				assert.Equal(t, fakescheme, scheme)
				assert.Equal(t, faketoken, token)
				// now the value should also be set in the authorization header
				scheme, token, err = ParseAuthorizationHeader(r)
				require.NoError(t, err)
				assert.Equal(t, fakescheme, scheme)
				assert.Equal(t, faketoken, token)
			})
		}
	})
	t.Run("valid authorisation header for websocket (workaround2)", func(t *testing.T) {
		r, err = http.NewRequest(http.MethodGet, faker.URL(), nil)
		require.NoError(t, err)
		r.Header.Add(HeaderWebsocketProtocol, fmt.Sprintf("%v, %v %v", headers.Authorization, fakescheme, faketoken))
		scheme, token, err = ParseAuthorizationHeader(r)
		require.NoError(t, err)
		assert.Equal(t, fakescheme, scheme)
		assert.Equal(t, faketoken, token)
		// now the value should also be set in the authorization header
		scheme, token, err = ParseAuthorizationHeader(r)
		require.NoError(t, err)
		assert.Equal(t, fakescheme, scheme)
		assert.Equal(t, faketoken, token)
	})
	t.Run("valid authorisation header for websocket (workaround3)", func(t *testing.T) {
		tokenString, err := GenerateAuthorizationHeaderValue(fakescheme, faketoken)
		require.NoError(t, err)
		tests := []struct {
			encoded string
		}{
			{
				encoded: base64.StdEncoding.EncodeToString([]byte(tokenString)),
			},
			{
				encoded: base64.URLEncoding.EncodeToString([]byte(tokenString)),
			},
			{
				encoded: base64.RawStdEncoding.EncodeToString([]byte(tokenString)),
			},
			{
				encoded: base64.RawURLEncoding.EncodeToString([]byte(tokenString)),
			},
		}
		for i := range tests {
			test := tests[i]
			t.Run("base64 encoding", func(t *testing.T) {
				r, err = http.NewRequest(http.MethodGet, faker.URL(), nil)
				require.NoError(t, err)
				r.Header.Add(HeaderWebsocketProtocol, fmt.Sprintf("%v, %v", headers.Authorization, test.encoded))
				scheme, token, err = ParseAuthorizationHeader(r)
				require.NoError(t, err)
				assert.Equal(t, fakescheme, scheme)
				assert.Equal(t, faketoken, token)
				// now the value should also be set in the authorization header
				scheme, token, err = ParseAuthorizationHeader(r)
				require.NoError(t, err)
				assert.Equal(t, fakescheme, scheme)
				assert.Equal(t, faketoken, token)
			})
		}
	})
}

func TestFromToRequestResponse(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, faker.URL(), nil)
	request.Header.Add(headers.Authorization, faker.Password())
	request.Header.Add(HeaderWebsocketProtocol, faker.Password())
	h := FromRequest(request)
	h.AppendHeader(headers.Accept, "1.0.0")
	h.AppendHeader(headers.AcceptEncoding, "gzip")
	r2 := httptest.NewRequest(http.MethodGet, faker.URL(), nil)
	assert.Empty(t, r2.Header)
	h.AppendToRequest(r2)
	assert.NotEmpty(t, r2.Header)
	h2 := FromRequest(r2)
	assert.True(t, h2.HasHeader(headers.Authorization))
	assert.True(t, h2.HasHeader(headers.AcceptEncoding))
	assert.True(t, h2.HasHeader(headers.Accept))
	assert.True(t, h2.HasHeader(HeaderWebsocketProtocol))

	response := httptest.NewRecorder()
	response.Header().Set(HeaderWebsocketProtocol, "base64.binary.k8s.io")
	response.Header().Set(headers.Authorization, faker.Password())
	h3 := FromResponse(response.Result())
	h3.AppendHeader(headers.Accept, "1.0.0")
	h3.AppendHeader(headers.AcceptEncoding, "gzip")
	response2 := httptest.NewRecorder()
	h3.AppendToResponse(response2)
	h4 := FromResponse(response2.Result())
	assert.True(t, h4.HasHeader(headers.Authorization))
	assert.True(t, h4.HasHeader(headers.AcceptEncoding))
	assert.True(t, h4.HasHeader(headers.Accept))
	assert.True(t, h4.HasHeader(HeaderWebsocketProtocol))
}

func TestAddProductInformationToUserAgent(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, faker.URL(), nil)
	require.NoError(t, err)
	assert.Empty(t, FetchUserAgent(r))
	require.NoError(t, AddProductInformationToUserAgent(r, faker.Word(), faker.IPv4(), ""))
	assert.NotEmpty(t, FetchUserAgent(r))
	require.NoError(t, AddProductInformationToUserAgent(r, faker.Word(), faker.IPv4(), faker.Sentence()))
	assert.NotEmpty(t, FetchUserAgent(r))
	fmt.Println(FetchUserAgent(r))
}

func TestSetLocationHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	assert.Empty(t, w.Header().Get(headers.Location))
	assert.Empty(t, w.Header().Get(headers.ContentLocation))
	location := faker.URL()
	SetLocationHeaders(w, location)
	assert.Equal(t, location, w.Header().Get(headers.Location))
	assert.Equal(t, location, w.Header().Get(headers.ContentLocation))
}

func TestGetHeaders(t *testing.T) {
	header := NewHeaders()
	test := faker.Word()
	header.AppendHeader(HeaderWebsocketProtocol, test)
	assert.Equal(t, test, header.Get(headers.Normalize(HeaderWebsocketProtocol))) //nolint:misspell
	assert.True(t, header.HasHeader(HeaderWebsocketProtocol))
	assert.True(t, header.HasHeader(headers.Normalize(HeaderWebsocketProtocol))) //nolint:misspell
	assert.Empty(t, header.Get(headers.ContentLocation))
	assert.False(t, header.HasHeader(headers.ContentLocation))
	assert.False(t, header.HasHeader(headers.Normalize(headers.ContentLocation))) //nolint:misspell
}

func TestSanitiseHeaders(t *testing.T) {
	header := &http.Header{}
	t.Run("empty", func(t *testing.T) {
		require.Empty(t, SanitiseHeaders(nil))
		require.Empty(t, SanitiseHeaders(header))
	})
	t.Run("valid", func(t *testing.T) {
		header.Add(headers.AcceptEncoding, "gzip")
		actual := SanitiseHeaders(header)
		require.NotNil(t, actual)
		assert.True(t, actual.HasHeader(
			headers.AcceptEncoding))
		header.Add(headers.Accept, "1.0.0")
		actual = SanitiseHeaders(header)
		assert.True(t, actual.HasHeader(
			headers.AcceptEncoding))
		assert.True(t, actual.HasHeader(
			headers.Accept))
	})
	t.Run("redact headers", func(t *testing.T) {
		header.Add(headers.Authorization, faker.Password())
		header.Add(HeaderWebsocketProtocol, faker.Password())
		actual := SanitiseHeaders(header)
		assert.True(t, actual.HasHeader(
			headers.AcceptEncoding))
		assert.True(t, actual.HasHeader(
			headers.Accept))
		assert.False(t, actual.HasHeader(
			headers.Authorization))
		assert.False(t, actual.HasHeader(
			HeaderWebsocketProtocol))
	})
	t.Run("allow/disallow list", func(t *testing.T) {
		h := NewHeaders()
		h.AppendHeader(headers.Authorization, faker.Password())
		h.AppendHeader(HeaderWebsocketProtocol, faker.Password())
		h.AppendHeader(headers.Accept, "1.0.0")
		h.AppendHeader(headers.AcceptEncoding, "gzip")
		h1 := h.Clone()
		h1.Sanitise()
		assert.True(t, h1.HasHeader(headers.Accept))
		assert.True(t, h1.HasHeader(headers.AcceptEncoding))
		assert.False(t, h1.HasHeader(HeaderWebsocketProtocol))
		assert.False(t, h1.HasHeader(headers.Authorization))
		assert.True(t, h.HasHeader(headers.Accept))
		assert.True(t, h.HasHeader(headers.AcceptEncoding))
		assert.True(t, h.HasHeader(HeaderWebsocketProtocol))
		assert.True(t, h.HasHeader(headers.Authorization))
		h11 := h.AllowList(headers.Authorization)
		assert.True(t, h11.HasHeader(headers.Accept))
		assert.True(t, h11.HasHeader(headers.AcceptEncoding))
		assert.False(t, h11.HasHeader(HeaderWebsocketProtocol))
		assert.True(t, h11.HasHeader(headers.Authorization))
		h2 := h.Clone()
		h2.Sanitise(headers.Authorization)
		h2.RemoveHeaders(headers.AcceptEncoding, headers.Accept)
		assert.False(t, h2.HasHeader(headers.Accept))
		assert.False(t, h2.HasHeader(headers.AcceptEncoding))
		assert.False(t, h2.HasHeader(HeaderWebsocketProtocol))
		assert.True(t, h2.HasHeader(headers.Authorization))
		h22 := h.DisallowList(headers.AcceptEncoding, headers.Accept)
		assert.False(t, h22.HasHeader(headers.Accept))
		assert.False(t, h22.HasHeader(headers.AcceptEncoding))
		assert.True(t, h22.HasHeader(HeaderWebsocketProtocol))
		assert.True(t, h22.HasHeader(headers.Authorization))
	})

}
