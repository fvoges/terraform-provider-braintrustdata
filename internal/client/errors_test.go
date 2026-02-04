package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAPIError_Structure verifies API error structure
func TestAPIError_Structure(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Message:    "Not found",
		Details:    map[string]interface{}{"resource": "project"},
	}

	if err.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", err.StatusCode)
	}
	if err.Message != "Not found" {
		t.Errorf("expected message 'Not found', got %q", err.Message)
	}
	if err.Details["resource"] != "project" {
		t.Errorf("expected resource=project in details, got %v", err.Details)
	}
}

// TestAPIError_Error verifies error message formatting
func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		Message:    "Invalid request",
	}

	expected := "API error: status 400, message: Invalid request"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

// TestDo_APIErrorHandling verifies API errors are parsed correctly
func TestDo_APIErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		expectedMsg    string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "404 not found",
			statusCode:     404,
			responseBody:   `{"error": "Resource not found"}`,
			expectedStatus: 404,
			expectedMsg:    "Resource not found",
		},
		{
			name:           "400 bad request",
			statusCode:     400,
			responseBody:   `{"error": "Invalid parameters"}`,
			expectedStatus: 400,
			expectedMsg:    "Invalid parameters",
		},
		{
			name:           "500 server error",
			statusCode:     500,
			responseBody:   `{"error": "Internal server error"}`,
			expectedStatus: 500,
			expectedMsg:    "Internal server error",
		},
		{
			name:           "non-JSON error response",
			statusCode:     403,
			responseBody:   `Forbidden`,
			expectedStatus: 403,
			expectedMsg:    "Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("sk-test", server.URL, "org-123")
			client.httpClient = server.Client()

			err := client.Do(context.Background(), "GET", "/test", nil, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			apiErr := &APIError{}
			ok := errors.As(err, &apiErr)
			if !ok {
				t.Fatalf("expected *APIError, got %T: %v", err, err)
			}

			if apiErr.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, apiErr.StatusCode)
			}
		})
	}
}

// TestDo_SensitiveDataSanitization verifies API keys are not leaked in errors
func TestDo_SensitiveDataSanitization(t *testing.T) {
	apiKey := "sk-secret-key-12345" //nolint:gosec // Test credential
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		// Echo the auth header in response (simulating a server error that leaks data)
		_, _ = w.Write([]byte(`{"error": "Invalid token: ` + r.Header.Get("Authorization") + `"}`))
	}))
	defer server.Close()

	client := NewClient(apiKey, server.URL, "org-123")
	client.httpClient = server.Client()

	err := client.Do(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errMsg := err.Error()

	// The error message should NOT contain the actual API key
	if containsSensitiveData(errMsg, apiKey) {
		t.Errorf("error message contains sensitive API key: %s", errMsg)
	}
}

// Helper function to check if error message contains sensitive data
func containsSensitiveData(msg, apiKey string) bool {
	// Check if the actual API key appears in the message
	// We expect "[REDACTED]" instead
	return len(apiKey) > 0 && len(msg) > 0 &&
		(contains(msg, apiKey) || contains(msg, "Bearer "+apiKey))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || contains(s[1:], substr)))
}
