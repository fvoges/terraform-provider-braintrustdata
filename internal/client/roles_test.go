package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRole(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateRoleRequest
		response       Role
		responseStatus int
		wantErr        bool
	}{
		{
			name: "creates role successfully",
			request: &CreateRoleRequest{
				Name:        "admin",
				Description: "Administrator role",
			},
			response: Role{
				ID:          "role-123",
				OrgID:       "org-456",
				Name:        "admin",
				Description: "Administrator role",
				Created:     "2024-01-15T10:30:00Z",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "handles missing required name",
			request: &CreateRoleRequest{
				Description: "No name provided",
			},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/role" {
					t.Errorf("expected path /v1/role, got %s", r.URL.Path)
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
			role, err := client.CreateRole(context.Background(), tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if role.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, role.ID)
			}
			if role.Name != tt.response.Name {
				t.Errorf("expected Name %s, got %s", tt.response.Name, role.Name)
			}
		})
	}
}

func TestGetRole(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		response       Role
		responseStatus int
		wantErr        bool
	}{
		{
			name:   "retrieves role successfully",
			roleID: "role-123",
			response: Role{
				ID:          "role-123",
				OrgID:       "org-456",
				Name:        "admin",
				Description: "Administrator role",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "handles not found",
			roleID:         "role-nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/role/" + tt.roleID
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
			role, err := client.GetRole(context.Background(), tt.roleID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if role.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, role.ID)
			}
		})
	}
}

func TestUpdateRole(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		request        *UpdateRoleRequest
		response       Role
		responseStatus int
		wantErr        bool
	}{
		{
			name:   "updates role successfully",
			roleID: "role-123",
			request: &UpdateRoleRequest{
				Name:        "updated-admin",
				Description: "Updated administrator role",
			},
			response: Role{
				ID:          "role-123",
				OrgID:       "org-456",
				Name:        "updated-admin",
				Description: "Updated administrator role",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/role/" + tt.roleID
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
			role, err := client.UpdateRole(context.Background(), tt.roleID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if role.Name != tt.response.Name {
				t.Errorf("expected Name %s, got %s", tt.response.Name, role.Name)
			}
		})
	}
}

func TestDeleteRole(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "deletes role successfully",
			roleID:         "role-123",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "handles not found",
			roleID:         "role-nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/role/" + tt.roleID
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
			err := client.DeleteRole(context.Background(), tt.roleID)

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

func TestListRoles(t *testing.T) {
	tests := []struct {
		options        *ListRolesOptions
		name           string
		expectedPath   string
		response       ListRolesResponse
		responseStatus int
		wantErr        bool
	}{
		{
			name:    "lists roles successfully",
			options: &ListRolesOptions{OrgName: "test-org"},
			response: ListRolesResponse{
				Roles: []Role{
					{
						ID:   "role-1",
						Name: "admin",
					},
					{
						ID:   "role-2",
						Name: "viewer",
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedPath:   "/v1/role?org_name=test-org",
		},
		{
			name: "lists roles with pagination",
			options: &ListRolesOptions{
				OrgName:       "test-org",
				Limit:         10,
				StartingAfter: "role-123",
			},
			response: ListRolesResponse{
				Roles: []Role{
					{ID: "role-2", Name: "viewer"},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedPath:   "/v1/role?org_name=test-org&limit=10&starting_after=role-123",
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
			result, err := client.ListRoles(context.Background(), tt.options)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(result.Roles) != len(tt.response.Roles) {
				t.Errorf("expected %d roles, got %d", len(tt.response.Roles), len(result.Roles))
			}
		})
	}
}
