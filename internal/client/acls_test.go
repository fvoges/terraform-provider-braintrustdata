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

func TestGetACL_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		aclID        string
		expectedPath string
		response     ACL
	}{
		{
			name:         "handles ID with slash",
			aclID:        "acl/123",
			expectedPath: "/v1/acl/acl/123",
			response: ACL{
				ID:       "acl/123",
				ObjectID: "project-123",
			},
		},
		{
			name:         "handles ID with space",
			aclID:        "acl 456",
			expectedPath: "/v1/acl/acl 456",
			response: ACL{
				ID:       "acl 456",
				ObjectID: "project-123",
			},
		},
		{
			name:         "handles ID with plus sign",
			aclID:        "acl+test",
			expectedPath: "/v1/acl/acl+test",
			response: ACL{
				ID:       "acl+test",
				ObjectID: "project-123",
			},
		},
		{
			name:         "handles ID with Unicode",
			aclID:        "アクセス",
			expectedPath: "/v1/acl/アクセス",
			response: ACL{
				ID:       "アクセス",
				ObjectID: "project-123",
			},
		},
		{
			name:         "handles ID with ampersand",
			aclID:        "acl&test",
			expectedPath: "/v1/acl/acl&test",
			response: ACL{
				ID:       "acl&test",
				ObjectID: "project-123",
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

			client := NewClient("sk-test", server.URL, "org-test")
			client.httpClient = server.Client()
			acl, err := client.GetACL(context.Background(), tt.aclID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if acl.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, acl.ID)
			}
		})
	}
}

func TestDeleteACL_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		aclID        string
		expectedPath string
	}{
		{
			name:         "handles ID with slash",
			aclID:        "acl/123",
			expectedPath: "/v1/acl/acl/123",
		},
		{
			name:         "handles ID with space",
			aclID:        "acl 456",
			expectedPath: "/v1/acl/acl 456",
		},
		{
			name:         "handles ID with plus sign",
			aclID:        "acl+test",
			expectedPath: "/v1/acl/acl+test",
		},
		{
			name:         "handles ID with Unicode",
			aclID:        "アクセス",
			expectedPath: "/v1/acl/アクセス",
		},
		{
			name:         "handles ID with ampersand",
			aclID:        "acl&test",
			expectedPath: "/v1/acl/acl&test",
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

			client := NewClient("sk-test", server.URL, "org-test")
			client.httpClient = server.Client()
			err := client.DeleteACL(context.Background(), tt.aclID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestListACLs_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		options      *ListACLsOptions
		expectedPath string
		response     ListACLsResponse
	}{
		{
			name: "handles object ID with slash",
			options: &ListACLsOptions{
				ObjectID:   "project/123",
				ObjectType: ACLObjectTypeProject,
			},
			expectedPath: "/v1/acl?object_id=project%2F123&object_type=project",
			response: ListACLsResponse{
				Objects: []ACL{
					{ID: "acl-1", ObjectID: "project/123"},
				},
			},
		},
		{
			name: "handles object ID with space",
			options: &ListACLsOptions{
				ObjectID:   "project 456",
				ObjectType: ACLObjectTypeProject,
			},
			expectedPath: "/v1/acl?object_id=project+456&object_type=project",
			response: ListACLsResponse{
				Objects: []ACL{
					{ID: "acl-2", ObjectID: "project 456"},
				},
			},
		},
		{
			name: "handles object ID with Unicode",
			options: &ListACLsOptions{
				ObjectID:   "プロジェクト",
				ObjectType: ACLObjectTypeProject,
			},
			expectedPath: "/v1/acl?object_id=%E3%83%97%E3%83%AD%E3%82%B8%E3%82%A7%E3%82%AF%E3%83%88&object_type=project",
			response: ListACLsResponse{
				Objects: []ACL{
					{ID: "acl-3", ObjectID: "プロジェクト"},
				},
			},
		},
		{
			name: "handles object ID with plus sign",
			options: &ListACLsOptions{
				ObjectID:   "project+test",
				ObjectType: ACLObjectTypeProject,
			},
			expectedPath: "/v1/acl?object_id=project%2Btest&object_type=project",
			response: ListACLsResponse{
				Objects: []ACL{
					{ID: "acl-4", ObjectID: "project+test"},
				},
			},
		},
		{
			name: "handles object ID with ampersand",
			options: &ListACLsOptions{
				ObjectID:   "project&test",
				ObjectType: ACLObjectTypeProject,
			},
			expectedPath: "/v1/acl?object_id=project%26test&object_type=project",
			response: ListACLsResponse{
				Objects: []ACL{
					{ID: "acl-5", ObjectID: "project&test"},
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

			client := NewClient("sk-test", server.URL, "org-test")
			client.httpClient = server.Client()
			result, err := client.ListACLs(context.Background(), tt.options)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(result.Objects) != len(tt.response.Objects) {
				t.Errorf("expected %d ACLs, got %d", len(tt.response.Objects), len(result.Objects))
			}
		})
	}
}
