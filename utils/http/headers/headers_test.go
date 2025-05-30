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

}
