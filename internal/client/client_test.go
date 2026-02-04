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

// TestNewClient_HTTPSOnlyEnforcement verifies that http:// URLs are rejected
func TestNewClient_HTTPSOnlyEnforcement(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "https URL accepted",
			baseURL: "https://api.braintrust.dev",
			wantErr: false,
		},
		{
			name:    "http URL rejected",
			baseURL: "http://api.braintrust.dev",
			wantErr: true,
		},
		{
			name:    "no scheme defaults to https",
			baseURL: "api.braintrust.dev",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("unexpected panic for URL %q: %v", tt.baseURL, r)
					}
				}
			}()

			client := NewClient("sk-test", tt.baseURL, "org-123")
			if tt.wantErr && client != nil {
				t.Errorf("expected error for URL %q, but client was created", tt.baseURL)
			}
		})
	}
}

// TestAuthHeader verifies Bearer token authentication is added correctly
func TestAuthHeader(t *testing.T) {
	apiKey := "sk-test-key-123" //nolint:gosec // Test credential
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + apiKey
		if authHeader != expectedAuth {
			t.Errorf("expected Authorization header %q, got %q", expectedAuth, authHeader)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient(apiKey, server.URL, "org-123")
	// Use the test server's http client which trusts its own cert
	client.httpClient = server.Client()

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	client.addAuthHeader(req)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Test code

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestUserAgent verifies user agent is set correctly
func TestUserAgent(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if !strings.HasPrefix(ua, "terraform-provider-braintrustdata/") {
			t.Errorf("expected User-Agent to start with terraform-provider-braintrustdata/, got %q", ua)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-123")
	client.httpClient = server.Client()

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	client.addAuthHeader(req)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Test code
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
