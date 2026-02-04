package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateAPIKeyRequest
		response       APIKey
		responseStatus int
		wantErr        bool
	}{
		{
			name: "creates api key successfully",
			request: &CreateAPIKeyRequest{
				Name:    "Test API Key",
				OrgName: "test-org",
			},
			response: APIKey{
				ID:          "apikey-123",
				OrgID:       "org-456",
				Name:        "Test API Key",
				PreviewName: "Test...Key",
				UserID:      "user-789",
				Created:     "2024-01-15T10:30:00Z",
				Key:         "sk-test-key-abc123",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "handles missing required name",
			request: &CreateAPIKeyRequest{
				OrgName: "test-org",
			},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/api_key" {
					t.Errorf("expected path /v1/api_key, got %s", r.URL.Path)
				}
				if r.Method != "POST" {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
				if !tt.wantErr {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			apiKey, err := client.CreateAPIKey(context.Background(), tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if apiKey.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, apiKey.ID)
			}
			if apiKey.Name != tt.response.Name {
				t.Errorf("expected Name %s, got %s", tt.response.Name, apiKey.Name)
			}
			if apiKey.Key != tt.response.Key {
				t.Errorf("expected Key %s, got %s", tt.response.Key, apiKey.Key)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		apiKeyID       string
		response       APIKey
		responseStatus int
		wantErr        bool
	}{
		{
			name:     "retrieves api key successfully",
			apiKeyID: "apikey-123",
			response: APIKey{
				ID:          "apikey-123",
				OrgID:       "org-456",
				Name:        "Test API Key",
				PreviewName: "Test...Key",
				UserID:      "user-789",
				Created:     "2024-01-15T10:30:00Z",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "handles not found",
			apiKeyID:       "apikey-nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/api_key/" + tt.apiKeyID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != "GET" {
					t.Errorf("expected GET method, got %s", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
				if !tt.wantErr {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			apiKey, err := client.GetAPIKey(context.Background(), tt.apiKeyID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if apiKey.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, apiKey.ID)
			}
		})
	}
}

func TestUpdateAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		apiKeyID       string
		request        *UpdateAPIKeyRequest
		response       APIKey
		responseStatus int
		wantErr        bool
	}{
		{
			name:     "updates api key successfully",
			apiKeyID: "apikey-123",
			request: &UpdateAPIKeyRequest{
				Name: "Updated API Key",
			},
			response: APIKey{
				ID:          "apikey-123",
				OrgID:       "org-456",
				Name:        "Updated API Key",
				PreviewName: "Updated...Key",
				UserID:      "user-789",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/api_key/" + tt.apiKeyID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != "PATCH" {
					t.Errorf("expected PATCH method, got %s", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			apiKey, err := client.UpdateAPIKey(context.Background(), tt.apiKeyID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if apiKey.Name != tt.response.Name {
				t.Errorf("expected Name %s, got %s", tt.response.Name, apiKey.Name)
			}
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		apiKeyID       string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "deletes api key successfully",
			apiKeyID:       "apikey-123",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "handles not found",
			apiKeyID:       "apikey-nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/api_key/" + tt.apiKeyID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != "DELETE" {
					t.Errorf("expected DELETE method, got %s", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			err := client.DeleteAPIKey(context.Background(), tt.apiKeyID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestListAPIKeys(t *testing.T) {
	tests := []struct {
		name           string
		options        *ListAPIKeysOptions
		expectedPath   string
		response       ListAPIKeysResponse
		responseStatus int
		wantErr        bool
	}{
		{
			name:    "lists api keys successfully",
			options: &ListAPIKeysOptions{OrgName: "test-org"},
			response: ListAPIKeysResponse{
				APIKeys: []APIKey{
					{
						ID:   "apikey-1",
						Name: "API Key 1",
					},
					{
						ID:   "apikey-2",
						Name: "API Key 2",
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedPath:   "/v1/api_key?org_name=test-org",
		},
		{
			name: "lists api keys with pagination",
			options: &ListAPIKeysOptions{
				OrgName:       "test-org",
				Limit:         10,
				StartingAfter: "apikey-123",
			},
			response: ListAPIKeysResponse{
				APIKeys: []APIKey{
					{ID: "apikey-2", Name: "API Key 2"},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedPath:   "/v1/api_key?org_name=test-org&limit=10&starting_after=apikey-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path+"?"+r.URL.RawQuery != tt.expectedPath {
					t.Errorf("expected path %s, got %s?%s", tt.expectedPath, r.URL.Path, r.URL.RawQuery)
				}
				if r.Method != "GET" {
					t.Errorf("expected GET method, got %s", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			result, err := client.ListAPIKeys(context.Background(), tt.options)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(result.APIKeys) != len(tt.response.APIKeys) {
				t.Errorf("expected %d api keys, got %d", len(tt.response.APIKeys), len(result.APIKeys))
			}
		})
	}
}
