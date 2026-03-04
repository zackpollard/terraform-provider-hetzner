// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const DefaultBaseURL = "https://robot-ws.your-server.de"

// APIError represents an error response from the Hetzner Robot API.
type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("hetzner api error (HTTP %d): %s - %s", e.StatusCode, e.ErrorCode, e.Message)
}

// apiErrorResponse matches the JSON error structure returned by the Hetzner Robot API.
type apiErrorResponse struct {
	Error struct {
		Status  int    `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Client is an HTTP client for the Hetzner Robot API.
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

// NewClient creates a new Hetzner Robot API client.
func NewClient(username, password string) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{},
	}
}

// Get performs a GET request to the given path.
func (c *Client) Get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post performs a POST request with URL-encoded form data.
func (c *Client) Post(path string, data url.Values) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, data)
}

// Put performs a PUT request with URL-encoded form data.
func (c *Client) Put(path string, data url.Values) ([]byte, error) {
	return c.doRequest(http.MethodPut, path, data)
}

// Delete performs a DELETE request to the given path.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}

// DeleteWithBody performs a DELETE request with URL-encoded form data.
func (c *Client) DeleteWithBody(path string, data url.Values) ([]byte, error) {
	return c.doRequest(http.MethodDelete, path, data)
}

// GetWithContext performs a GET request with context support.
func (c *Client) GetWithContext(ctx context.Context, path string) ([]byte, error) {
	return c.doRequestWithContext(ctx, http.MethodGet, path, nil)
}

// PostWithContext performs a POST request with context support.
func (c *Client) PostWithContext(ctx context.Context, path string, data url.Values) ([]byte, error) {
	return c.doRequestWithContext(ctx, http.MethodPost, path, data)
}

// doRequest executes an HTTP request with basic auth and handles error responses.
func (c *Client) doRequest(method, path string, data url.Values) ([]byte, error) {
	return c.doRequestWithContext(context.Background(), method, path, data)
}

// doRequestWithContext executes an HTTP request with context and basic auth, and handles error responses.
func (c *Client) doRequestWithContext(ctx context.Context, method, path string, data url.Values) ([]byte, error) {
	reqURL := c.BaseURL + path

	var body io.Reader
	if data != nil {
		body = strings.NewReader(data.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")

	if data != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}

		var errResp apiErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Code != "" {
			apiErr.ErrorCode = errResp.Error.Code
			apiErr.Message = errResp.Error.Message
		} else {
			apiErr.ErrorCode = "UNKNOWN"
			apiErr.Message = string(respBody)
		}

		return nil, apiErr
	}

	return respBody, nil
}
