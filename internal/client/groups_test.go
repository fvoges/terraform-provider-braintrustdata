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

// TestCreateGroup verifies group creation
func TestCreateGroup(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/group" {
			t.Errorf("expected path /v1/group, got %s", r.URL.Path)
		}

		var req CreateGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name == "" {
			t.Error("expected name to be set")
		}

		resp := Group{
			ID:          "group-123",
			Name:        req.Name,
			OrgID:       req.OrgID,
			Description: req.Description,
			Created:     time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	group, err := client.CreateGroup(context.Background(), &CreateGroupRequest{
		Name:        "Test Group",
		OrgID:       "org-test",
		Description: "A test group",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if group.ID != "group-123" {
		t.Errorf("expected ID group-123, got %s", group.ID)
	}
	if group.Name != "Test Group" {
		t.Errorf("expected name 'Test Group', got %s", group.Name)
	}
	// Note: member_ids are not returned during creation as the API doesn't accept them
	// Members must be added via a separate update call
}

// TestGetGroup verifies group retrieval
func TestGetGroup(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/group/group-123" {
			t.Errorf("expected path /v1/group/group-123, got %s", r.URL.Path)
		}

		resp := Group{
			ID:          "group-123",
			Name:        "Test Group",
			OrgID:       "org-test",
			Description: "A test group",
			MemberIDs:   []string{"user-1", "user-2"},
			Created:     time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	group, err := client.GetGroup(context.Background(), "group-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if group.ID != "group-123" {
		t.Errorf("expected ID group-123, got %s", group.ID)
	}
	if group.Name != "Test Group" {
		t.Errorf("expected name 'Test Group', got %s", group.Name)
	}
}

// TestGetGroup_NotFound verifies 404 handling
func TestGetGroup_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Group not found",
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetGroup(context.Background(), "nonexistent")

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

// TestUpdateGroup verifies group updates
func TestUpdateGroup(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/group/group-123" {
			t.Errorf("expected path /v1/group/group-123, got %s", r.URL.Path)
		}

		var req UpdateGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		resp := Group{
			ID:          "group-123",
			Name:        req.Name,
			Description: req.Description,
			MemberIDs:   req.MemberIDs,
			OrgID:       "org-test",
			Created:     time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	group, err := client.UpdateGroup(context.Background(), "group-123", &UpdateGroupRequest{
		Name:        "Updated Group",
		Description: "Updated description",
		MemberIDs:   []string{"user-1", "user-2", "user-3"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if group.Name != "Updated Group" {
		t.Errorf("expected name 'Updated Group', got %s", group.Name)
	}
	if len(group.MemberIDs) != 3 {
		t.Errorf("expected 3 member IDs, got %d", len(group.MemberIDs))
	}
}

// TestDeleteGroup verifies group deletion
func TestDeleteGroup(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/group/group-123" {
			t.Errorf("expected path /v1/group/group-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "group-123",
			"deleted": true,
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteGroup(context.Background(), "group-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestListGroups verifies group listing
func TestListGroups(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/group" {
			t.Errorf("expected path /v1/group, got %s", r.URL.Path)
		}

		// Check query parameters
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			t.Error("expected org_id query parameter")
		}

		resp := ListGroupsResponse{
			Groups: []Group{
				{
					ID:      "group-1",
					Name:    "Group 1",
					OrgID:   orgID,
					Created: time.Now().Format(time.RFC3339),
				},
				{
					ID:      "group-2",
					Name:    "Group 2",
					OrgID:   orgID,
					Created: time.Now().Format(time.RFC3339),
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListGroups(context.Background(), &ListGroupsOptions{
		OrgID: "org-test",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(result.Groups))
	}

	if result.Groups[0].Name != "Group 1" {
		t.Errorf("expected first group name 'Group 1', got %s", result.Groups[0].Name)
	}
}
