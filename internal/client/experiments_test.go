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

// TestCreateExperiment verifies experiment creation
func TestCreateExperiment(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/experiment" {
			t.Errorf("expected path /v1/experiment, got %s", r.URL.Path)
		}

		var req CreateExperimentRequest
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

		resp := Experiment{
			ID:          "experiment-123",
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
		"framework": "pytorch",
		"version":   "2.0",
	}

	publicTrue := true
	experiment, err := client.CreateExperiment(context.Background(), &CreateExperimentRequest{
		ProjectID:   "project-123",
		Name:        "Test Experiment",
		Description: "A test experiment",
		Public:      &publicTrue,
		Metadata:    metadata,
		Tags:        []string{"ml", "test"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if experiment.ID != "experiment-123" {
		t.Errorf("expected ID experiment-123, got %s", experiment.ID)
	}
	if experiment.Name != "Test Experiment" {
		t.Errorf("expected name 'Test Experiment', got %s", experiment.Name)
	}
	if experiment.ProjectID != "project-123" {
		t.Errorf("expected project_id 'project-123', got %s", experiment.ProjectID)
	}
	if !experiment.Public {
		t.Error("expected public to be true")
	}
	if len(experiment.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(experiment.Tags))
	}
}

// TestGetExperiment verifies experiment retrieval
func TestGetExperiment(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/experiment/experiment-123" {
			t.Errorf("expected path /v1/experiment/experiment-123, got %s", r.URL.Path)
		}

		resp := Experiment{
			ID:          "experiment-123",
			ProjectID:   "project-123",
			Name:        "Test Experiment",
			Description: "A test experiment",
			Public:      true,
			Metadata: map[string]interface{}{
				"framework": "pytorch",
			},
			Tags:    []string{"ml", "test"},
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

	experiment, err := client.GetExperiment(context.Background(), "experiment-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if experiment.ID != "experiment-123" {
		t.Errorf("expected ID experiment-123, got %s", experiment.ID)
	}
	if experiment.Name != "Test Experiment" {
		t.Errorf("expected name 'Test Experiment', got %s", experiment.Name)
	}
	if experiment.ProjectID != "project-123" {
		t.Errorf("expected project_id 'project-123', got %s", experiment.ProjectID)
	}
}

// TestGetExperiment_NotFound verifies 404 handling
func TestGetExperiment_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Experiment not found",
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetExperiment(context.Background(), "nonexistent")

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

// TestUpdateExperiment verifies experiment updates
func TestUpdateExperiment(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/experiment/experiment-123" {
			t.Errorf("expected path /v1/experiment/experiment-123, got %s", r.URL.Path)
		}

		var req UpdateExperimentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Test that Public pointer works correctly
		public := false
		if req.Public != nil {
			public = *req.Public
		}

		// Handle metadata pointer
		var metadata map[string]interface{}
		if req.Metadata != nil {
			metadata = *req.Metadata
		}

		resp := Experiment{
			ID:          "experiment-123",
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
	experiment, err := client.UpdateExperiment(context.Background(), "experiment-123", &UpdateExperimentRequest{
		Name:        "Updated Experiment",
		Description: "Updated description",
		Public:      &publicFalse,
		Metadata:    &metadata,
		Tags:        []string{"updated"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if experiment.Name != "Updated Experiment" {
		t.Errorf("expected name 'Updated Experiment', got %s", experiment.Name)
	}
	if experiment.Public {
		t.Error("expected public to be false")
	}
	if len(experiment.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(experiment.Tags))
	}
}

// TestDeleteExperiment verifies experiment deletion
func TestDeleteExperiment(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/experiment/experiment-123" {
			t.Errorf("expected path /v1/experiment/experiment-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "experiment-123",
			"deleted": true,
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteExperiment(context.Background(), "experiment-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDeleteExperiment_NotFound verifies 404 handling for delete (idempotency)
func TestDeleteExperiment_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Experiment not found",
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteExperiment(context.Background(), "nonexistent-id")

	// Should return NotFoundError
	if !IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got: %v", err)
	}
}

// TestListExperiments verifies experiment listing
func TestListExperiments(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/experiment" {
			t.Errorf("expected path /v1/experiment, got %s", r.URL.Path)
		}

		// Check query parameters
		projectID := r.URL.Query().Get("project_id")
		if projectID == "" {
			t.Error("expected project_id query parameter")
		}

		resp := ListExperimentsResponse{
			Experiments: []Experiment{
				{
					ID:        "experiment-1",
					ProjectID: projectID,
					Name:      "Experiment 1",
					Created:   time.Now().Format(time.RFC3339),
				},
				{
					ID:        "experiment-2",
					ProjectID: projectID,
					Name:      "Experiment 2",
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

	result, err := client.ListExperiments(context.Background(), &ListExperimentsOptions{
		ProjectID: "project-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Experiments) != 2 {
		t.Errorf("expected 2 experiments, got %d", len(result.Experiments))
	}

	if result.Experiments[0].Name != "Experiment 1" {
		t.Errorf("expected first experiment name 'Experiment 1', got %s", result.Experiments[0].Name)
	}
}

// TestGetExperiment_EmptyID verifies empty ID validation
func TestGetExperiment_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.GetExperiment(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyExperimentID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyExperimentID, err)
	}
}

// TestUpdateExperiment_EmptyID verifies empty ID validation
func TestUpdateExperiment_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.UpdateExperiment(context.Background(), "", &UpdateExperimentRequest{
		Name: "Test",
	})

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyExperimentID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyExperimentID, err)
	}
}

// TestDeleteExperiment_EmptyID verifies empty ID validation
func TestDeleteExperiment_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	err := client.DeleteExperiment(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyExperimentID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyExperimentID, err)
	}
}

// TestListExperiments_WithOptions verifies query parameter handling
func TestListExperiments_WithOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *ListExperimentsOptions
		wantPath string
	}{
		{
			name:     "with project_id only",
			opts:     &ListExperimentsOptions{ProjectID: "project-123"},
			wantPath: "/v1/experiment?project_id=project-123",
		},
		{
			name:     "with limit only",
			opts:     &ListExperimentsOptions{Limit: 10},
			wantPath: "/v1/experiment?limit=10",
		},
		{
			name:     "with cursor only",
			opts:     &ListExperimentsOptions{Cursor: "cursor-abc"},
			wantPath: "/v1/experiment?cursor=cursor-abc",
		},
		{
			name:     "with all options",
			opts:     &ListExperimentsOptions{ProjectID: "project-123", Limit: 10, Cursor: "cursor-abc"},
			wantPath: "/v1/experiment?cursor=cursor-abc&limit=10&project_id=project-123",
		},
		{
			name:     "with nil options",
			opts:     nil,
			wantPath: "/v1/experiment",
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

				resp := ListExperimentsResponse{
					Experiments: []Experiment{},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient("sk-test", server.URL, "org-test")
			client.httpClient = server.Client()

			_, err := client.ListExperiments(context.Background(), tt.opts)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
