// Package hc is a small wrapper around http.Client that provides a few
// convenience methods.
package hc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"oss.acmcsuf.com/qg/backend/qg"
)

// Client wraps an HTTP client.
type Client struct {
	http.Client
	baseURL url.URL
}

// NewClient returns a new client.
func NewClient(baseURL string, c *http.Client) *Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	return &Client{*c, *u}
}

// NewDefaultClient returns a new client with the default HTTP client.
func NewDefaultClient(baseURL string) *Client {
	return NewClient(baseURL, http.DefaultClient)
}

// GET performs a GET request.
func (c *Client) GET(ctx context.Context, path string, q url.Values, respBody any) error {
	u := c.baseURL
	u.Path = path
	u.RawQuery = q.Encode()
	return c.do(ctx, http.MethodGet, u, nil, respBody)
}

// GET performs a GET request. It is a type-safe wrapper around Client.GET.
func GET[T any](ctx context.Context, c *Client, path string, q url.Values) (*T, error) {
	var respBody T
	if err := c.GET(ctx, path, q, &respBody); err != nil {
		return nil, err
	}
	return &respBody, nil
}

// POST performs a POST request.
func (c *Client) POST(ctx context.Context, path string, reqBody, respBody any) error {
	u := c.baseURL
	u.Path = path
	return c.do(ctx, http.MethodPost, u, reqBody, respBody)
}

// POST performs a POST request. It is a type-safe wrapper around Client.POST.
func POST[T any](ctx context.Context, c *Client, path string, reqBody any) (*T, error) {
	var respBody T
	if err := c.POST(ctx, path, reqBody, &respBody); err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (c *Client) do(ctx context.Context, method string, u url.URL, reqBody, respBody any) error {
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	req.Body = io.NopCloser(bytes.NewReader(reqBodyJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var err qg.Error
		if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
			return errors.Errorf("unexpected status code %d (no error in body)", resp.StatusCode)
		}
		return errors.Errorf("server returned error status %d: %s", resp.StatusCode, err.Message)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	return nil
}
