package client

import (
	"context"
	"encoding/json"
	"errors"
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
			expectedPath:   "/v1/api_key?limit=10&org_name=test-org&starting_after=apikey-123",
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

func TestGetAPIKey_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		apiKeyID     string
		expectedPath string
		response     APIKey
	}{
		{
			name:         "handles ID with slash",
			apiKeyID:     "apikey/123",
			expectedPath: "/v1/api_key/apikey/123",
			response: APIKey{
				ID:   "apikey/123",
				Name: "Test API Key",
			},
		},
		{
			name:         "handles ID with space",
			apiKeyID:     "apikey 456",
			expectedPath: "/v1/api_key/apikey 456",
			response: APIKey{
				ID:   "apikey 456",
				Name: "Test API Key",
			},
		},
		{
			name:         "handles ID with plus sign",
			apiKeyID:     "apikey+test",
			expectedPath: "/v1/api_key/apikey+test",
			response: APIKey{
				ID:   "apikey+test",
				Name: "Test API Key",
			},
		},
		{
			name:         "handles ID with Unicode",
			apiKeyID:     "キー",
			expectedPath: "/v1/api_key/キー",
			response: APIKey{
				ID:   "キー",
				Name: "Test API Key",
			},
		},
		{
			name:         "handles ID with ampersand",
			apiKeyID:     "apikey&test",
			expectedPath: "/v1/api_key/apikey&test",
			response: APIKey{
				ID:   "apikey&test",
				Name: "Test API Key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}
				if r.Method != "GET" {
					t.Errorf("expected GET method, got %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			apiKey, err := client.GetAPIKey(context.Background(), tt.apiKeyID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if apiKey.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, apiKey.ID)
			}
		})
	}
}

func TestUpdateAPIKey_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		apiKeyID     string
		expectedPath string
		request      *UpdateAPIKeyRequest
		response     APIKey
	}{
		{
			name:         "handles ID with slash",
			apiKeyID:     "apikey/123",
			expectedPath: "/v1/api_key/apikey/123",
			request: &UpdateAPIKeyRequest{
				Name: "Updated API Key",
			},
			response: APIKey{
				ID:   "apikey/123",
				Name: "Updated API Key",
			},
		},
		{
			name:         "handles ID with space",
			apiKeyID:     "apikey 456",
			expectedPath: "/v1/api_key/apikey 456",
			request: &UpdateAPIKeyRequest{
				Name: "Updated API Key",
			},
			response: APIKey{
				ID:   "apikey 456",
				Name: "Updated API Key",
			},
		},
		{
			name:         "handles ID with plus sign",
			apiKeyID:     "apikey+test",
			expectedPath: "/v1/api_key/apikey+test",
			request: &UpdateAPIKeyRequest{
				Name: "Updated API Key",
			},
			response: APIKey{
				ID:   "apikey+test",
				Name: "Updated API Key",
			},
		},
		{
			name:         "handles ID with Unicode",
			apiKeyID:     "キー",
			expectedPath: "/v1/api_key/キー",
			request: &UpdateAPIKeyRequest{
				Name: "Updated API Key",
			},
			response: APIKey{
				ID:   "キー",
				Name: "Updated API Key",
			},
		},
		{
			name:         "handles ID with ampersand",
			apiKeyID:     "apikey&test",
			expectedPath: "/v1/api_key/apikey&test",
			request: &UpdateAPIKeyRequest{
				Name: "Updated API Key",
			},
			response: APIKey{
				ID:   "apikey&test",
				Name: "Updated API Key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}
				if r.Method != "PATCH" {
					t.Errorf("expected PATCH method, got %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			apiKey, err := client.UpdateAPIKey(context.Background(), tt.apiKeyID, tt.request)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if apiKey.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, apiKey.ID)
			}
		})
	}
}

func TestDeleteAPIKey_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		apiKeyID     string
		expectedPath string
	}{
		{
			name:         "handles ID with slash",
			apiKeyID:     "apikey/123",
			expectedPath: "/v1/api_key/apikey/123",
		},
		{
			name:         "handles ID with space",
			apiKeyID:     "apikey 456",
			expectedPath: "/v1/api_key/apikey 456",
		},
		{
			name:         "handles ID with plus sign",
			apiKeyID:     "apikey+test",
			expectedPath: "/v1/api_key/apikey+test",
		},
		{
			name:         "handles ID with Unicode",
			apiKeyID:     "キー",
			expectedPath: "/v1/api_key/キー",
		},
		{
			name:         "handles ID with ampersand",
			apiKeyID:     "apikey&test",
			expectedPath: "/v1/api_key/apikey&test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("expected path %s, got %s", tt.expectedPath, r.URL.Path)
				}
				if r.Method != "DELETE" {
					t.Errorf("expected DELETE method, got %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			err := client.DeleteAPIKey(context.Background(), tt.apiKeyID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestListAPIKeys_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		options      *ListAPIKeysOptions
		expectedPath string
		response     ListAPIKeysResponse
	}{
		{
			name: "handles org name with space",
			options: &ListAPIKeysOptions{
				OrgName: "test org",
			},
			expectedPath: "/v1/api_key?org_name=test+org",
			response: ListAPIKeysResponse{
				APIKeys: []APIKey{
					{ID: "apikey-1", Name: "API Key 1"},
				},
			},
		},
		{
			name: "handles starting_after with special characters",
			options: &ListAPIKeysOptions{
				OrgName:       "test-org",
				StartingAfter: "apikey/123",
			},
			expectedPath: "/v1/api_key?org_name=test-org&starting_after=apikey%2F123",
			response: ListAPIKeysResponse{
				APIKeys: []APIKey{
					{ID: "apikey-2", Name: "API Key 2"},
				},
			},
		},
		{
			name: "handles api key name with Unicode",
			options: &ListAPIKeysOptions{
				OrgName:    "test-org",
				APIKeyName: "キー",
			},
			expectedPath: "/v1/api_key?api_key_name=%E3%82%AD%E3%83%BC&org_name=test-org",
			response: ListAPIKeysResponse{
				APIKeys: []APIKey{
					{ID: "apikey-3", Name: "キー"},
				},
			},
		},
		{
			name: "handles plus sign in parameters",
			options: &ListAPIKeysOptions{
				OrgName:    "test-org",
				APIKeyName: "apikey+test",
			},
			expectedPath: "/v1/api_key?api_key_name=apikey%2Btest&org_name=test-org",
			response: ListAPIKeysResponse{
				APIKeys: []APIKey{
					{ID: "apikey-4", Name: "apikey+test"},
				},
			},
		},
		{
			name: "handles ampersand in org name",
			options: &ListAPIKeysOptions{
				OrgName: "test&org",
			},
			expectedPath: "/v1/api_key?org_name=test%26org",
			response: ListAPIKeysResponse{
				APIKeys: []APIKey{
					{ID: "apikey-5", Name: "API Key 5"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fullPath := r.URL.Path + "?" + r.URL.RawQuery
				if fullPath != tt.expectedPath {
					t.Errorf("expected path %s, got %s", tt.expectedPath, fullPath)
				}
				if r.Method != "GET" {
					t.Errorf("expected GET method, got %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			result, err := client.ListAPIKeys(context.Background(), tt.options)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(result.APIKeys) != len(tt.response.APIKeys) {
				t.Errorf("expected %d api keys, got %d", len(tt.response.APIKeys), len(result.APIKeys))
			}
		})
	}
}

// TestGetAPIKey_EmptyID verifies empty ID validation
func TestGetAPIKey_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.GetAPIKey(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyAPIKeyID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyAPIKeyID, err)
	}
}

// TestUpdateAPIKey_EmptyID verifies empty ID validation
func TestUpdateAPIKey_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.UpdateAPIKey(context.Background(), "", &UpdateAPIKeyRequest{Name: "test"})

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyAPIKeyID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyAPIKeyID, err)
	}
}

// TestDeleteAPIKey_EmptyID verifies empty ID validation
func TestDeleteAPIKey_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	err := client.DeleteAPIKey(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyAPIKeyID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyAPIKeyID, err)
	}
}
