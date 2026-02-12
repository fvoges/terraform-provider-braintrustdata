package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCreateDataset verifies dataset creation
func TestCreateDataset(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/dataset" {
			t.Errorf("expected path /v1/dataset, got %s", r.URL.Path)
		}

		var req CreateDatasetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name == "" {
			t.Error("expected name to be set")
		}
		if req.ProjectID == "" {
			t.Error("expected project_id to be set")
		}

		// Handle Public pointer
		public := false
		if req.Public != nil {
			public = *req.Public
		}

		resp := Dataset{
			ID:          "dataset-123",
			ProjectID:   req.ProjectID,
			Name:        req.Name,
			Description: req.Description,
			Public:      public,
			Metadata:    req.Metadata,
			Tags:        req.Tags,
			Created:     time.Now().Format(time.RFC3339),
			UserID:      "user-123",
			OrgID:       "org-test",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	metadata := map[string]interface{}{
		"source": "test-suite",
		"size":   float64(100),
	}

	publicTrue := true
	dataset, err := client.CreateDataset(context.Background(), &CreateDatasetRequest{
		ProjectID:   "project-123",
		Name:        "Test Dataset",
		Description: "A test dataset",
		Public:      &publicTrue,
		Metadata:    metadata,
		Tags:        []string{"images", "test"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dataset.ID != "dataset-123" {
		t.Errorf("expected ID dataset-123, got %s", dataset.ID)
	}
	if dataset.Name != "Test Dataset" {
		t.Errorf("expected name 'Test Dataset', got %s", dataset.Name)
	}
	if dataset.ProjectID != "project-123" {
		t.Errorf("expected project_id 'project-123', got %s", dataset.ProjectID)
	}
	if !dataset.Public {
		t.Error("expected public to be true")
	}
	if len(dataset.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(dataset.Tags))
	}
}

// TestGetDataset verifies dataset retrieval
func TestGetDataset(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/dataset/dataset-123" {
			t.Errorf("expected path /v1/dataset/dataset-123, got %s", r.URL.Path)
		}

		resp := Dataset{
			ID:          "dataset-123",
			ProjectID:   "project-123",
			Name:        "Test Dataset",
			Description: "A test dataset",
			Public:      true,
			Metadata: map[string]interface{}{
				"source": "test-suite",
			},
			Tags:    []string{"images", "test"},
			Created: time.Now().Format(time.RFC3339),
			UserID:  "user-123",
			OrgID:   "org-test",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	dataset, err := client.GetDataset(context.Background(), "dataset-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dataset.ID != "dataset-123" {
		t.Errorf("expected ID dataset-123, got %s", dataset.ID)
	}
	if dataset.Name != "Test Dataset" {
		t.Errorf("expected name 'Test Dataset', got %s", dataset.Name)
	}
	if dataset.ProjectID != "project-123" {
		t.Errorf("expected project_id 'project-123', got %s", dataset.ProjectID)
	}
}

// TestGetDataset_NotFound verifies 404 handling
func TestGetDataset_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Dataset not found",
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetDataset(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr := &APIError{}
	ok := errors.As(err, &apiErr)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

// TestUpdateDataset verifies dataset updates
func TestUpdateDataset(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/dataset/dataset-123" {
			t.Errorf("expected path /v1/dataset/dataset-123, got %s", r.URL.Path)
		}

		var req UpdateDatasetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Test that Public pointer works correctly
		public := false
		if req.Public != nil {
			public = *req.Public
		}

		// Handle metadata
		metadata := req.Metadata

		resp := Dataset{
			ID:          "dataset-123",
			ProjectID:   "project-123",
			Name:        req.Name,
			Description: req.Description,
			Public:      public,
			Metadata:    metadata,
			Tags:        req.Tags,
			Created:     time.Now().Format(time.RFC3339),
			UserID:      "user-123",
			OrgID:       "org-test",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	publicFalse := false
	metadata := map[string]interface{}{
		"updated": true,
	}
	dataset, err := client.UpdateDataset(context.Background(), "dataset-123", &UpdateDatasetRequest{
		Name:        "Updated Dataset",
		Description: "Updated description",
		Public:      &publicFalse,
		Metadata:    metadata,
		Tags:        []string{"updated"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dataset.Name != "Updated Dataset" {
		t.Errorf("expected name 'Updated Dataset', got %s", dataset.Name)
	}
	if dataset.Public {
		t.Error("expected public to be false")
	}
	if len(dataset.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(dataset.Tags))
	}
}

// TestDeleteDataset verifies dataset deletion
func TestDeleteDataset(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/dataset/dataset-123" {
			t.Errorf("expected path /v1/dataset/dataset-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "dataset-123",
			"deleted": true,
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteDataset(context.Background(), "dataset-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDeleteDataset_NotFound verifies 404 handling for delete (idempotency)
func TestDeleteDataset_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Dataset not found",
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteDataset(context.Background(), "nonexistent-id")

	// Should return NotFoundError
	if !IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got: %v", err)
	}
}

// TestListDatasets verifies dataset listing
func TestListDatasets(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/dataset" {
			t.Errorf("expected path /v1/dataset, got %s", r.URL.Path)
		}

		// Check query parameters
		projectID := r.URL.Query().Get("project_id")
		if projectID == "" {
			t.Error("expected project_id query parameter")
		}

		resp := ListDatasetsResponse{
			Datasets: []Dataset{
				{
					ID:        "dataset-1",
					ProjectID: projectID,
					Name:      "Dataset 1",
					Created:   time.Now().Format(time.RFC3339),
				},
				{
					ID:        "dataset-2",
					ProjectID: projectID,
					Name:      "Dataset 2",
					Created:   time.Now().Format(time.RFC3339),
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListDatasets(context.Background(), &ListDatasetsOptions{
		ProjectID: "project-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Datasets) != 2 {
		t.Errorf("expected 2 datasets, got %d", len(result.Datasets))
	}

	if result.Datasets[0].Name != "Dataset 1" {
		t.Errorf("expected first dataset name 'Dataset 1', got %s", result.Datasets[0].Name)
	}
}

// TestGetDataset_EmptyID verifies empty ID validation
func TestGetDataset_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.GetDataset(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyDatasetID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyDatasetID, err)
	}
}

// TestUpdateDataset_EmptyID verifies empty ID validation
func TestUpdateDataset_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.UpdateDataset(context.Background(), "", &UpdateDatasetRequest{
		Name: "Test",
	})

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyDatasetID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyDatasetID, err)
	}
}

// TestDeleteDataset_EmptyID verifies empty ID validation
func TestDeleteDataset_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	err := client.DeleteDataset(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyDatasetID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyDatasetID, err)
	}
}

// TestListDatasets_WithOptions verifies query parameter handling
func TestListDatasets_WithOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *ListDatasetsOptions
		wantPath string
	}{
		{
			name:     "with project_id only",
			opts:     &ListDatasetsOptions{ProjectID: "project-123"},
			wantPath: "/v1/dataset?project_id=project-123",
		},
		{
			name:     "with limit only",
			opts:     &ListDatasetsOptions{Limit: 10},
			wantPath: "/v1/dataset?limit=10",
		},
		{
			name:     "with cursor only",
			opts:     &ListDatasetsOptions{Cursor: "cursor-abc"},
			wantPath: "/v1/dataset?cursor=cursor-abc",
		},
		{
			name:     "with all options",
			opts:     &ListDatasetsOptions{ProjectID: "project-123", Limit: 10, Cursor: "cursor-abc"},
			wantPath: "/v1/dataset?cursor=cursor-abc&limit=10&project_id=project-123",
		},
		{
			name:     "with nil options",
			opts:     nil,
			wantPath: "/v1/dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fullPath := r.URL.Path
				if r.URL.RawQuery != "" {
					fullPath += "?" + r.URL.RawQuery
				}

				if fullPath != tt.wantPath {
					t.Errorf("expected path %s, got %s", tt.wantPath, fullPath)
				}

				resp := ListDatasetsResponse{
					Datasets: []Dataset{},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient("sk-test", server.URL, "org-test")
			client.httpClient = server.Client()

			_, err := client.ListDatasets(context.Background(), tt.opts)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestGetDataset_SpecialCharacters verifies URL path escaping for IDs with special characters
func TestGetDataset_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name           string
		datasetID      string
		expectedPath   string
		response       Dataset
		responseStatus int
	}{
		{
			name:         "handles ID with space",
			datasetID:    "dataset 123",
			expectedPath: "/v1/dataset/dataset 123",
			response: Dataset{
				ID:   "dataset 123",
				Name: "Test Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with slash",
			datasetID:    "dataset/123",
			expectedPath: "/v1/dataset/dataset/123",
			response: Dataset{
				ID:   "dataset/123",
				Name: "Test Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with question mark",
			datasetID:    "dataset?123",
			expectedPath: "/v1/dataset/dataset?123",
			response: Dataset{
				ID:   "dataset?123",
				Name: "Test Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with hash",
			datasetID:    "dataset#123",
			expectedPath: "/v1/dataset/dataset#123",
			response: Dataset{
				ID:   "dataset#123",
				Name: "Test Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with percent",
			datasetID:    "dataset%123",
			expectedPath: "/v1/dataset/dataset%123",
			response: Dataset{
				ID:   "dataset%123",
				Name: "Test Dataset",
			},
			responseStatus: http.StatusOK,
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

				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			dataset, err := client.GetDataset(context.Background(), tt.datasetID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if dataset.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, dataset.ID)
			}
		})
	}
}

// TestUpdateDataset_SpecialCharacters verifies URL path escaping for IDs with special characters
func TestUpdateDataset_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name           string
		datasetID      string
		expectedPath   string
		request        *UpdateDatasetRequest
		response       Dataset
		responseStatus int
	}{
		{
			name:         "handles ID with space",
			datasetID:    "dataset 123",
			expectedPath: "/v1/dataset/dataset 123",
			request: &UpdateDatasetRequest{
				Name: "Updated Dataset",
			},
			response: Dataset{
				ID:   "dataset 123",
				Name: "Updated Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with slash",
			datasetID:    "dataset/123",
			expectedPath: "/v1/dataset/dataset/123",
			request: &UpdateDatasetRequest{
				Name: "Updated Dataset",
			},
			response: Dataset{
				ID:   "dataset/123",
				Name: "Updated Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with question mark",
			datasetID:    "dataset?123",
			expectedPath: "/v1/dataset/dataset?123",
			request: &UpdateDatasetRequest{
				Name: "Updated Dataset",
			},
			response: Dataset{
				ID:   "dataset?123",
				Name: "Updated Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with hash",
			datasetID:    "dataset#123",
			expectedPath: "/v1/dataset/dataset#123",
			request: &UpdateDatasetRequest{
				Name: "Updated Dataset",
			},
			response: Dataset{
				ID:   "dataset#123",
				Name: "Updated Dataset",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with percent",
			datasetID:    "dataset%123",
			expectedPath: "/v1/dataset/dataset%123",
			request: &UpdateDatasetRequest{
				Name: "Updated Dataset",
			},
			response: Dataset{
				ID:   "dataset%123",
				Name: "Updated Dataset",
			},
			responseStatus: http.StatusOK,
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

				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			dataset, err := client.UpdateDataset(context.Background(), tt.datasetID, tt.request)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if dataset.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, dataset.ID)
			}
		})
	}
}

// TestDeleteDataset_SpecialCharacters verifies URL path escaping for IDs with special characters
func TestDeleteDataset_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name           string
		datasetID      string
		expectedPath   string
		responseStatus int
	}{
		{
			name:           "handles ID with space",
			datasetID:      "dataset 123",
			expectedPath:   "/v1/dataset/dataset 123",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with slash",
			datasetID:      "dataset/123",
			expectedPath:   "/v1/dataset/dataset/123",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with question mark",
			datasetID:      "dataset?123",
			expectedPath:   "/v1/dataset/dataset?123",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with hash",
			datasetID:      "dataset#123",
			expectedPath:   "/v1/dataset/dataset#123",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with percent",
			datasetID:      "dataset%123",
			expectedPath:   "/v1/dataset/dataset%123",
			responseStatus: http.StatusOK,
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

				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL, "test-org")
			client.httpClient = server.Client()
			err := client.DeleteDataset(context.Background(), tt.datasetID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
