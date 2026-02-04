// Package client provides HTTP client functionality for the Braintrust API.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Version is the provider version, used in User-Agent header
const Version = "0.1.0"

// Client is the API client for Braintrust
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	orgID      string
	userAgent  string
}

// NewClient creates a new Braintrust API client
// baseURL must use https:// or have no scheme (defaults to https)
// Panics if baseURL uses http:// for security
func NewClient(apiKey, baseURL, orgID string) *Client {
	// Enforce HTTPS only
	if strings.HasPrefix(strings.ToLower(baseURL), "http://") {
		panic("http:// URLs are not allowed, must use https:// for security")
	}

	// Add https:// if no scheme provided
	if !strings.HasPrefix(strings.ToLower(baseURL), "https://") {
		baseURL = "https://" + baseURL
	}

	// Validate it's a proper URL
	_, err := url.Parse(baseURL)
	if err != nil {
		panic(fmt.Sprintf("invalid baseURL: %v", err))
	}

	// Configure HTTP client with TLS 1.2+ and timeout
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // 0x0303
		},
	}

	httpClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: httpClient,
		apiKey:     apiKey,
		orgID:      orgID,
		userAgent:  fmt.Sprintf("terraform-provider-braintrustdata/%s", Version),
	}
}

// addAuthHeader adds the Bearer token authentication header
func (c *Client) addAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
}

// OrgID returns the client's organization ID
func (c *Client) OrgID() string {
	return c.orgID
}

// Do executes an HTTP request with the given method, path, body, and response destination
func (c *Client) Do(ctx context.Context, method, path string, body, v interface{}) error {
	// Build URL
	fullURL := c.baseURL + path

	// Marshal body if provided
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.addAuthHeader(req)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	// Unmarshal response if destination provided
	if v != nil {
		if err := json.Unmarshal(respBody, v); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
