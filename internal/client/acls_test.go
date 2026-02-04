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

// TestCreateACL verifies ACL creation
func TestCreateACL(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/acl" {
			t.Errorf("expected path /v1/acl, got %s", r.URL.Path)
		}

		var req CreateACLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.ObjectID == "" {
			t.Error("expected object_id to be set")
		}
		if req.ObjectType == "" {
			t.Error("expected object_type to be set")
		}

		resp := ACL{
			ID:         "acl-123",
			ObjectID:   req.ObjectID,
			ObjectType: req.ObjectType,
			UserID:     req.UserID,
			GroupID:    req.GroupID,
			RoleID:     req.RoleID,
			Permission: req.Permission,
			Created:    time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	acl, err := client.CreateACL(context.Background(), &CreateACLRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
		UserID:     "user-456",
		Permission: PermissionRead,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acl.ID != "acl-123" {
		t.Errorf("expected ID acl-123, got %s", acl.ID)
	}
	if acl.ObjectID != "project-123" {
		t.Errorf("expected object_id project-123, got %s", acl.ObjectID)
	}
	if acl.ObjectType != ACLObjectTypeProject {
		t.Errorf("expected object_type project, got %s", acl.ObjectType)
	}
}

// TestGetACL verifies ACL retrieval
func TestGetACL(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/acl/acl-123" {
			t.Errorf("expected path /v1/acl/acl-123, got %s", r.URL.Path)
		}

		resp := ACL{
			ID:         "acl-123",
			ObjectID:   "project-123",
			ObjectType: ACLObjectTypeProject,
			UserID:     "user-456",
			Permission: PermissionRead,
			Created:    time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	acl, err := client.GetACL(context.Background(), "acl-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acl.ID != "acl-123" {
		t.Errorf("expected ID acl-123, got %s", acl.ID)
	}
	if acl.ObjectID != "project-123" {
		t.Errorf("expected object_id project-123, got %s", acl.ObjectID)
	}
}

// TestGetACL_NotFound verifies 404 handling
func TestGetACL_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "ACL not found",
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetACL(context.Background(), "nonexistent")

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

// TestDeleteACL verifies ACL deletion
func TestDeleteACL(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/acl/acl-123" {
			t.Errorf("expected path /v1/acl/acl-123, got %s", r.URL.Path)
		}

		resp := ACL{
			ID:         "acl-123",
			ObjectID:   "project-123",
			ObjectType: ACLObjectTypeProject,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteACL(context.Background(), "acl-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestListACLs verifies ACL listing
func TestListACLs(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/acl" {
			t.Errorf("expected path /v1/acl, got %s", r.URL.Path)
		}

		// Check required query parameters
		objectID := r.URL.Query().Get("object_id")
		objectType := r.URL.Query().Get("object_type")
		if objectID == "" {
			t.Error("expected object_id query parameter")
		}
		if objectType == "" {
			t.Error("expected object_type query parameter")
		}

		resp := ListACLsResponse{
			Objects: []ACL{
				{
					ID:         "acl-1",
					ObjectID:   objectID,
					ObjectType: ACLObjectType(objectType),
					UserID:     "user-1",
					Permission: PermissionRead,
					Created:    time.Now().Format(time.RFC3339),
				},
				{
					ID:         "acl-2",
					ObjectID:   objectID,
					ObjectType: ACLObjectType(objectType),
					GroupID:    "group-1",
					Permission: PermissionUpdate,
					Created:    time.Now().Format(time.RFC3339),
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListACLs(context.Background(), &ListACLsOptions{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Objects) != 2 {
		t.Errorf("expected 2 ACLs, got %d", len(result.Objects))
	}

	if result.Objects[0].ID != "acl-1" {
		t.Errorf("expected first ACL ID 'acl-1', got %s", result.Objects[0].ID)
	}
	if result.Objects[0].UserID != "user-1" {
		t.Errorf("expected first ACL user_id 'user-1', got %s", result.Objects[0].UserID)
	}
}
