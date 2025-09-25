package http

import (
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	headers2 "github.com/go-http-utils/headers"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/http/headers"
)

type ClientWithHeaders struct {
	client  IClient
	headers headers.Headers
}

func newClientWithHeaders(underlyingClient IClient, headerValues ...string) (c *ClientWithHeaders, err error) {
	c = &ClientWithHeaders{
		headers: make(headers.Headers),
	}

	if underlyingClient == nil {
		c.client = NewPlainHTTPClient()
	} else {
		c.client = underlyingClient
	}

	for header := range slices.Chunk(headerValues, 2) {
		if len(header) != 2 {
			err = commonerrors.New(commonerrors.ErrInvalid, "headers must be supplied in key-value pairs")
			return
		}

		c.headers.AppendHeader(header[0], header[1])
	}

	return
}

func NewHTTPClientWithHeaders(headers ...string) (clientWithHeaders IClientWithHeaders, err error) {
	return newClientWithHeaders(nil, headers...)
}

func NewHTTPClientWithEmptyHeaders() (c IClientWithHeaders, err error) {
	return NewHTTPClientWithHeaders()
}

func NewHTTPClientWithUnderlyingClientWithHeaders(underlyingClient IClient, headers ...string) (c IClientWithHeaders, err error) {
	return newClientWithHeaders(underlyingClient, headers...)
}

func NewHTTPClientWithUnderlyingClientWithEmptyHeaders(underlyingClient IClient) (c IClientWithHeaders, err error) {
	return newClientWithHeaders(underlyingClient)
}

func (c *ClientWithHeaders) do(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *ClientWithHeaders) Head(url string) (*http.Response, error) {
	return c.do(http.MethodHead, url, nil)
}

func (c *ClientWithHeaders) Post(url, contentType string, rawBody interface{}) (*http.Response, error) {
	b, err := determineBodyReader(rawBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headers2.ContentType, contentType) // make sure to overwrite any in the headers
	return c.client.Do(req)
}

func (c *ClientWithHeaders) PostForm(url string, data url.Values) (*http.Response, error) {
	rawBody := strings.NewReader(data.Encode())
	return c.Post(url, headers.MIMEXWWWFormURLEncoded, rawBody)
}

func (c *ClientWithHeaders) StandardClient() *http.Client {
	return c.client.StandardClient()
}

func (c *ClientWithHeaders) Get(url string) (*http.Response, error) {
	return c.do(http.MethodGet, url, nil)
}

func (c *ClientWithHeaders) Do(req *http.Request) (*http.Response, error) {
	c.headers.AppendToRequest(req)
	return c.client.Do(req)
}

func (c *ClientWithHeaders) Delete(url string) (*http.Response, error) {
	return c.do(http.MethodDelete, url, nil)
}

func (c *ClientWithHeaders) Put(url string, rawBody interface{}) (*http.Response, error) {
	b, err := determineBodyReader(rawBody)
	if err != nil {
		return nil, err
	}
	return c.do(http.MethodPut, url, b)
}

func (c *ClientWithHeaders) Options(url string) (*http.Response, error) {
	return c.do(http.MethodOptions, url, nil)
}

func (c *ClientWithHeaders) Close() error {
	c.client.StandardClient().CloseIdleConnections()
	return nil
}

func (c *ClientWithHeaders) AppendHeader(key, value string) {
	if c.headers == nil {
		c.headers = make(headers.Headers)
	}
	c.headers.AppendHeader(key, value)
}

func (c *ClientWithHeaders) RemoveHeader(key string) {
	if c.headers == nil {
		return
	}
	delete(c.headers, key)
}

func (c *ClientWithHeaders) ClearHeaders() {
	c.headers = make(headers.Headers)
}
