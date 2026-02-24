package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewClient verifies basic client creation with required fields
func TestNewClient(t *testing.T) {
	apiKey := "sk-test-key"
	baseURL := "https://api.braintrust.dev"
	orgID := "org-123"

	client := NewClient(apiKey, baseURL, orgID)

	if client == nil {
		t.Fatal("expected client to be created, got nil")
	}
	if client.apiKey != apiKey {
		t.Errorf("expected apiKey %q, got %q", apiKey, client.apiKey)
	}
	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %q, got %q", baseURL, client.baseURL)
	}
	if client.orgID != orgID {
		t.Errorf("expected orgID %q, got %q", orgID, client.orgID)
	}
	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestNewClient_DoesNotPanicOnInvalidBaseURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "http scheme",
			url:  "http://api.braintrust.dev",
		},
		{
			name: "userinfo in URL",
			url:  "https://user@api.braintrust.dev",
		},
		{
			name: "empty URL",
			url:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("unexpected panic for URL %q: %v", tt.url, r)
				}
			}()

			_ = NewClient("sk-test", tt.url, "org-123")
		})
	}
}

func TestNewClient_NormalizesBaseURLWithoutScheme(t *testing.T) {
	client := NewClient("sk-test", "api.braintrust.dev/", "org-123")
	if client.baseURL != "https://api.braintrust.dev" {
		t.Fatalf("expected normalized baseURL to be https://api.braintrust.dev, got %q", client.baseURL)
	}
}

// TestAuthHeader verifies Bearer token authentication is added correctly
func TestAuthHeader(t *testing.T) {
	apiKey := "sk-test-key-123" //nolint:gosec // Test credential
	client := NewClient(apiKey, "https://api.braintrust.dev", "org-123")

	req, err := http.NewRequest("GET", "https://api.braintrust.dev/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	client.addAuthHeader(req)
	authHeader := req.Header.Get("Authorization")
	expectedAuth := "Bearer " + apiKey
	if authHeader != expectedAuth {
		t.Errorf("expected Authorization header %q, got %q", expectedAuth, authHeader)
	}
}

// TestUserAgent verifies user agent is set correctly
func TestUserAgent(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-123")
	req, err := http.NewRequest("GET", "https://api.braintrust.dev/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	client.addAuthHeader(req)
	ua := req.Header.Get("User-Agent")
	if !strings.HasPrefix(ua, "terraform-provider-braintrustdata/") {
		t.Errorf("expected User-Agent to start with terraform-provider-braintrustdata/, got %q", ua)
	}
}

// TestHTTPTimeout verifies timeout is configured
func TestHTTPTimeout(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-123")

	if client.httpClient.Timeout == 0 {
		t.Error("expected httpClient.Timeout to be configured, got 0")
	}

	expectedTimeout := 60 * time.Second
	if client.httpClient.Timeout != expectedTimeout {
		t.Errorf("expected timeout %v, got %v", expectedTimeout, client.httpClient.Timeout)
	}
}

// TestTLSConfiguration verifies TLS 1.2+ enforcement
func TestTLSConfiguration(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-123")

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected httpClient.Transport to be *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("expected TLSClientConfig to be set")
	}

	// MinVersion should be TLS 1.2 (0x0303)
	expectedMinVersion := uint16(0x0303) // TLS 1.2
	if transport.TLSClientConfig.MinVersion != expectedMinVersion {
		t.Errorf("expected MinVersion %d (TLS 1.2), got %d", expectedMinVersion, transport.TLSClientConfig.MinVersion)
	}
}

// TestDo_ContextCancellation verifies context cancellation is handled
func TestDo_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-123")
	client.httpClient = server.Client()

	// Create a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Do(ctx, "GET", "/test", nil, nil)
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got: %v", err)
	}
}

// TestDo_Success verifies successful request/response handling
func TestDo_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": "test-123", "name": "Test"}`))
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-123")
	client.httpClient = server.Client()

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	err := client.Do(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "test-123" {
		t.Errorf("expected ID test-123, got %s", result.ID)
	}
	if result.Name != "Test" {
		t.Errorf("expected Name Test, got %s", result.Name)
	}
}

func TestDo_RejectsAbsoluteURLPath(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-123")

	err := client.Do(context.Background(), "GET", "https://example.com/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for absolute request URL, got nil")
	}
	if !strings.Contains(err.Error(), "absolute request URLs are not allowed") {
		t.Fatalf("expected absolute URL error, got %v", err)
	}
}

func TestDo_AllowsQueryOnlyRelativePath(t *testing.T) {
	var gotPath string
	var gotRawQuery string
	var gotRequestURI string

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery
		gotRequestURI = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-123")
	client.httpClient = server.Client()

	err := client.Do(context.Background(), "GET", "?a=b", nil, nil)
	if err != nil {
		t.Fatalf("expected query-only relative path to be accepted, got error: %v", err)
	}
	if gotPath != "/" {
		t.Fatalf("expected normalized path '/', got %q", gotPath)
	}
	if gotRawQuery != "a=b" {
		t.Fatalf("expected raw query %q, got %q", "a=b", gotRawQuery)
	}
	if gotRequestURI != "/?a=b" {
		t.Fatalf("expected request URI %q, got %q", "/?a=b", gotRequestURI)
	}
}

func TestDo_RejectsInvalidBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		errMatch string
	}{
		{
			name:     "rejects http base URL",
			baseURL:  "http://api.braintrust.dev",
			errMatch: "only https is allowed",
		},
		{
			name:     "rejects empty host",
			baseURL:  "https://",
			errMatch: "host cannot be empty",
		},
		{
			name:     "rejects user info",
			baseURL:  "https://user@api.braintrust.dev",
			errMatch: "user info is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("sk-test", tt.baseURL, "org-123")

			err := client.Do(context.Background(), "GET", "/test", nil, nil)
			if err == nil {
				t.Fatalf("expected error for baseURL %q, got nil", tt.baseURL)
			}
			if !strings.Contains(err.Error(), tt.errMatch) {
				t.Fatalf("expected error containing %q, got %v", tt.errMatch, err)
			}
		})
	}
}

func TestDo_RejectsUnsafeRelativePaths(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-123")

	tests := []struct {
		name     string
		path     string
		errMatch string
	}{
		{
			name:     "rejects non-rooted path",
			path:     "test",
			errMatch: "must be rooted",
		},
		{
			name:     "rejects dot segment traversal",
			path:     "/v1/../secret",
			errMatch: "must not contain dot segments",
		},
		{
			name:     "rejects encoded dot segment traversal",
			path:     "/v1/%2e%2e/secret",
			errMatch: "must not contain dot segments",
		},
		{
			name:     "rejects scheme-relative absolute URL",
			path:     "//example.com/secret",
			errMatch: "absolute request URLs are not allowed",
		},
		{
			name:     "rejects fragment in path",
			path:     "/v1/test#fragment",
			errMatch: "must not contain fragments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Do(context.Background(), "GET", tt.path, nil, nil)
			if err == nil {
				t.Fatalf("expected error for path %q, got nil", tt.path)
			}
			if !strings.Contains(err.Error(), tt.errMatch) {
				t.Fatalf("expected error containing %q, got %v", tt.errMatch, err)
			}
		})
	}
}
