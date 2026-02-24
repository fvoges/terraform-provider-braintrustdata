// Package client provides HTTP client functionality for the Braintrust API.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
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
func NewClient(apiKey, baseURL, orgID string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" && !strings.Contains(baseURL, "://") {
		baseURL = "https://" + baseURL
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

func validateBaseURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("base URL cannot be empty")
	}
	baseURL, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid client base URL: %w", err)
	}
	if !strings.EqualFold(baseURL.Scheme, "https") {
		return nil, fmt.Errorf("insecure base URL scheme %q: only https is allowed", baseURL.Scheme)
	}
	if baseURL.Host == "" {
		return nil, errors.New("base URL host cannot be empty")
	}
	if baseURL.User != nil {
		return nil, errors.New("base URL user info is not allowed")
	}
	return baseURL, nil
}

func validateRequestPath(raw string) (*url.URL, error) {
	pathURL, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid request path: %w", err)
	}
	if pathURL.IsAbs() || pathURL.Host != "" || pathURL.Scheme != "" {
		return nil, fmt.Errorf("absolute request URLs are not allowed: %q", raw)
	}
	if pathURL.Fragment != "" {
		return nil, fmt.Errorf("request path must not contain fragments: %q", raw)
	}
	if pathURL.Path == "" {
		pathURL.Path = "/"
		pathURL.RawPath = ""
	}
	if !strings.HasPrefix(pathURL.Path, "/") {
		return nil, fmt.Errorf("request path must be rooted (start with '/'): %q", raw)
	}

	decodedPath, err := url.PathUnescape(pathURL.EscapedPath())
	if err != nil {
		return nil, fmt.Errorf("invalid request path encoding: %w", err)
	}
	for _, segment := range strings.Split(decodedPath, "/") {
		if segment == "." || segment == ".." {
			return nil, fmt.Errorf("request path must not contain dot segments: %q", raw)
		}
	}

	return pathURL, nil
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
	baseURL, err := validateBaseURL(c.baseURL)
	if err != nil {
		return err
	}
	pathURL, err := validateRequestPath(path)
	if err != nil {
		return err
	}
	fullURL := baseURL.ResolveReference(pathURL).String()

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
	resp, err := c.httpClient.Do(req) //nolint:gosec // G704 false positive: baseURL/path are validated above (https + relative path only).
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
