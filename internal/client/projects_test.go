package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateProjectRequest
		response       Project
		responseStatus int
		wantErr        bool
	}{
		{
			name: "creates project successfully",
			request: &CreateProjectRequest{
				Name:        "Test Project",
				Description: "A test project",
			},
			response: Project{
				ID:          "proj-123",
				OrgID:       "org-456",
				Name:        "Test Project",
				Description: "A test project",
				Created:     "2024-01-15T10:30:00Z",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "handles missing required name",
			request: &CreateProjectRequest{
				Description: "No name provided",
			},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/project" {
					t.Errorf("expected path /v1/project, got %s", r.URL.Path)
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
			project, err := client.CreateProject(context.Background(), tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if project.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, project.ID)
			}
			if project.Name != tt.response.Name {
				t.Errorf("expected Name %s, got %s", tt.response.Name, project.Name)
			}
		})
	}
}

func TestGetProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		response       Project
		responseStatus int
		wantErr        bool
	}{
		{
			name:      "retrieves project successfully",
			projectID: "proj-123",
			response: Project{
				ID:          "proj-123",
				OrgID:       "org-456",
				Name:        "Test Project",
				Description: "A test project",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "handles not found",
			projectID:      "proj-nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/project/" + tt.projectID
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
			project, err := client.GetProject(context.Background(), tt.projectID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if project.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, project.ID)
			}
		})
	}
}

func TestUpdateProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		request        *UpdateProjectRequest
		response       Project
		responseStatus int
		wantErr        bool
	}{
		{
			name:      "updates project successfully",
			projectID: "proj-123",
			request: &UpdateProjectRequest{
				Name:        "Updated Project",
				Description: "Updated description",
			},
			response: Project{
				ID:          "proj-123",
				OrgID:       "org-456",
				Name:        "Updated Project",
				Description: "Updated description",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/project/" + tt.projectID
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
			project, err := client.UpdateProject(context.Background(), tt.projectID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if project.Name != tt.response.Name {
				t.Errorf("expected Name %s, got %s", tt.response.Name, project.Name)
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "deletes project successfully",
			projectID:      "proj-123",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "handles not found",
			projectID:      "proj-nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/v1/project/" + tt.projectID
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
			err := client.DeleteProject(context.Background(), tt.projectID)

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

func TestListProjects(t *testing.T) {
	tests := []struct {
		options        *ListProjectsOptions
		name           string
		expectedPath   string
		response       ListProjectsResponse
		responseStatus int
		wantErr        bool
	}{
		{
			name:    "lists projects successfully",
			options: &ListProjectsOptions{OrgName: "test-org"},
			response: ListProjectsResponse{
				Projects: []Project{
					{
						ID:   "proj-1",
						Name: "Project 1",
					},
					{
						ID:   "proj-2",
						Name: "Project 2",
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedPath:   "/v1/project?org_name=test-org",
		},
		{
			name: "lists projects with pagination",
			options: &ListProjectsOptions{
				OrgName:       "test-org",
				Limit:         10,
				StartingAfter: "proj-123",
			},
			response: ListProjectsResponse{
				Projects: []Project{
					{ID: "proj-2", Name: "Project 2"},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedPath:   "/v1/project?limit=10&org_name=test-org&starting_after=proj-123",
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
			result, err := client.ListProjects(context.Background(), tt.options)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(result.Projects) != len(tt.response.Projects) {
				t.Errorf("expected %d projects, got %d", len(tt.response.Projects), len(result.Projects))
			}
		})
	}
}

func TestGetProject_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		expectedPath   string
		response       Project
		responseStatus int
	}{
		{
			name:         "handles ID with slash",
			projectID:    "project/123",
			expectedPath: "/v1/project/project/123",
			response: Project{
				ID:   "project/123",
				Name: "Test Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with space",
			projectID:    "project 456",
			expectedPath: "/v1/project/project 456",
			response: Project{
				ID:   "project 456",
				Name: "Test Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with plus sign",
			projectID:    "project+test",
			expectedPath: "/v1/project/project+test",
			response: Project{
				ID:   "project+test",
				Name: "Test Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with Unicode",
			projectID:    "プロジェクト",
			expectedPath: "/v1/project/プロジェクト",
			response: Project{
				ID:   "プロジェクト",
				Name: "Test Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with ampersand",
			projectID:    "project&test",
			expectedPath: "/v1/project/project&test",
			response: Project{
				ID:   "project&test",
				Name: "Test Project",
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
			project, err := client.GetProject(context.Background(), tt.projectID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if project.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, project.ID)
			}
		})
	}
}

func TestUpdateProject_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		expectedPath   string
		request        *UpdateProjectRequest
		response       Project
		responseStatus int
	}{
		{
			name:         "handles ID with slash",
			projectID:    "project/123",
			expectedPath: "/v1/project/project/123",
			request: &UpdateProjectRequest{
				Name: "Updated Project",
			},
			response: Project{
				ID:   "project/123",
				Name: "Updated Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with space",
			projectID:    "project 456",
			expectedPath: "/v1/project/project 456",
			request: &UpdateProjectRequest{
				Name: "Updated Project",
			},
			response: Project{
				ID:   "project 456",
				Name: "Updated Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with plus sign",
			projectID:    "project+test",
			expectedPath: "/v1/project/project+test",
			request: &UpdateProjectRequest{
				Name: "Updated Project",
			},
			response: Project{
				ID:   "project+test",
				Name: "Updated Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with Unicode",
			projectID:    "プロジェクト",
			expectedPath: "/v1/project/プロジェクト",
			request: &UpdateProjectRequest{
				Name: "Updated Project",
			},
			response: Project{
				ID:   "プロジェクト",
				Name: "Updated Project",
			},
			responseStatus: http.StatusOK,
		},
		{
			name:         "handles ID with ampersand",
			projectID:    "project&test",
			expectedPath: "/v1/project/project&test",
			request: &UpdateProjectRequest{
				Name: "Updated Project",
			},
			response: Project{
				ID:   "project&test",
				Name: "Updated Project",
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
			project, err := client.UpdateProject(context.Background(), tt.projectID, tt.request)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if project.ID != tt.response.ID {
				t.Errorf("expected ID %s, got %s", tt.response.ID, project.ID)
			}
		})
	}
}

func TestDeleteProject_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		expectedPath   string
		responseStatus int
	}{
		{
			name:           "handles ID with slash",
			projectID:      "project/123",
			expectedPath:   "/v1/project/project/123",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with space",
			projectID:      "project 456",
			expectedPath:   "/v1/project/project 456",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with plus sign",
			projectID:      "project+test",
			expectedPath:   "/v1/project/project+test",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with Unicode",
			projectID:      "プロジェクト",
			expectedPath:   "/v1/project/プロジェクト",
			responseStatus: http.StatusOK,
		},
		{
			name:           "handles ID with ampersand",
			projectID:      "project&test",
			expectedPath:   "/v1/project/project&test",
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
			err := client.DeleteProject(context.Background(), tt.projectID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestListProjects_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		options      *ListProjectsOptions
		expectedPath string
		response     ListProjectsResponse
	}{
		{
			name: "handles org name with space",
			options: &ListProjectsOptions{
				OrgName: "test org",
			},
			expectedPath: "/v1/project?org_name=test+org",
			response: ListProjectsResponse{
				Projects: []Project{
					{ID: "proj-1", Name: "Project 1"},
				},
			},
		},
		{
			name: "handles starting_after with special characters",
			options: &ListProjectsOptions{
				OrgName:       "test-org",
				StartingAfter: "project/123",
			},
			expectedPath: "/v1/project?org_name=test-org&starting_after=project%2F123",
			response: ListProjectsResponse{
				Projects: []Project{
					{ID: "proj-2", Name: "Project 2"},
				},
			},
		},
		{
			name: "handles project name with Unicode",
			options: &ListProjectsOptions{
				OrgName:     "test-org",
				ProjectName: "プロジェクト",
			},
			expectedPath: "/v1/project?org_name=test-org&project_name=%E3%83%97%E3%83%AD%E3%82%B8%E3%82%A7%E3%82%AF%E3%83%88",
			response: ListProjectsResponse{
				Projects: []Project{
					{ID: "proj-3", Name: "プロジェクト"},
				},
			},
		},
		{
			name: "handles plus sign in parameters",
			options: &ListProjectsOptions{
				OrgName:     "test-org",
				ProjectName: "project+test",
			},
			expectedPath: "/v1/project?org_name=test-org&project_name=project%2Btest",
			response: ListProjectsResponse{
				Projects: []Project{
					{ID: "proj-4", Name: "project+test"},
				},
			},
		},
		{
			name: "handles ampersand in org name",
			options: &ListProjectsOptions{
				OrgName: "test&org",
			},
			expectedPath: "/v1/project?org_name=test%26org",
			response: ListProjectsResponse{
				Projects: []Project{
					{ID: "proj-5", Name: "Project 5"},
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
			result, err := client.ListProjects(context.Background(), tt.options)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(result.Projects) != len(tt.response.Projects) {
				t.Errorf("expected %d projects, got %d", len(tt.response.Projects), len(result.Projects))
			}
		})
	}
}

// TestGetProject_EmptyID verifies empty ID validation
func TestGetProject_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.GetProject(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyProjectID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyProjectID, err)
	}
}

// TestUpdateProject_EmptyID verifies empty ID validation
func TestUpdateProject_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	_, err := client.UpdateProject(context.Background(), "", &UpdateProjectRequest{Name: "test"})

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyProjectID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyProjectID, err)
	}
}

// TestDeleteProject_EmptyID verifies empty ID validation
func TestDeleteProject_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.example.com", "org-test")

	err := client.DeleteProject(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}

	if !errors.Is(err, ErrEmptyProjectID) {
		t.Errorf("expected error '%v', got '%v'", ErrEmptyProjectID, err)
	}
}
