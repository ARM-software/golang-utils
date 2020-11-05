package http

import (
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
)

// IClient provides a familiar HTTP client interface with automatic retries and exponential backoff.
type IClient interface {
	// Get is a convenience helper for doing simple GET requests.
	Get(url string) (*http.Response, error)
	// Head is a convenience method for doing simple HEAD requests.
	Head(url string) (*http.Response, error)
	// Post is a convenience method for doing simple POST requests.
	Post(url, bodyType string, body interface{}) (*http.Response, error)
	// PostForm is a convenience method for doing simple POST operations using
	// pre-filled url.Values form data.
	PostForm(url string, data url.Values) (*http.Response, error)
	// StandardClient returns a stdlib *http.Client with a custom Transport, which
	// shims in a *retryablehttp.Client for added retries.
	StandardClient() *http.Client
	// Perform a PUT request.
	Put(url string, rawBody interface{}) (*http.Response, error)
	// Perform a DELETE request.
	Delete(url string) (*http.Response, error)
	// Perform a generic request with exponential backoff
	Do(req *http.Request) (*http.Response, error)
}

type RetryableClient struct {
	client *retryablehttp.Client
}

func NewRetryableClient() IClient {
	return &RetryableClient{client: retryablehttp.NewClient()}
}

func (c *RetryableClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

func (c *RetryableClient) Post(url, bodyType string, body interface{}) (*http.Response, error) {
	return c.client.Post(url, bodyType, body)
}

func (c *RetryableClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.client.PostForm(url, data)
}

func (c *RetryableClient) StandardClient() *http.Client {
	return c.client.StandardClient()
}

func (c *RetryableClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	r, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}
	return c.client.Do(r)
}

func (c *RetryableClient) Delete(url string) (*http.Response, error) {
	req, err := retryablehttp.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *RetryableClient) Put(url string, rawBody interface{}) (*http.Response, error) {
	req, err := retryablehttp.NewRequest("PUT", url, rawBody)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}
