package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetTag(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/project_tag/tag-123" {
			t.Errorf("expected path /v1/project_tag/tag-123, got %s", r.URL.Path)
		}

		resp := Tag{
			ID:        "tag-123",
			Name:      "production",
			ProjectID: "proj-123",
			UserID:    "user-123",
			Color:     "#00AAFF",
			Created:   "2026-02-26T10:30:00Z",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	tag, err := client.GetTag(context.Background(), "tag-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tag.ID != "tag-123" {
		t.Errorf("expected id tag-123, got %s", tag.ID)
	}
	if tag.Name != "production" {
		t.Errorf("expected name production, got %s", tag.Name)
	}
}

func TestGetTag_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetTag(context.Background(), "")
	if !errors.Is(err, ErrEmptyTagID) {
		t.Fatalf("expected ErrEmptyTagID, got %v", err)
	}
}

func TestListTags_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/project_tag" {
			t.Errorf("expected path /v1/project_tag, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("limit"); got != "10" {
			t.Errorf("expected limit 10, got %q", got)
		}
		if got := query.Get("starting_after"); got != "cursor-next" {
			t.Errorf("expected starting_after cursor-next, got %q", got)
		}
		if got := query.Get("ending_before"); got != "cursor-prev" {
			t.Errorf("expected ending_before cursor-prev, got %q", got)
		}
		if got := query.Get("org_name"); got != "test-org" {
			t.Errorf("expected org_name test-org, got %q", got)
		}
		if got := query.Get("project_id"); got != "proj-123" {
			t.Errorf("expected project_id proj-123, got %q", got)
		}
		if got := query.Get("project_name"); got != "example-project" {
			t.Errorf("expected project_name example-project, got %q", got)
		}
		if got := query.Get("project_tag_name"); got != "production" {
			t.Errorf("expected project_tag_name production, got %q", got)
		}
		if got := query["ids"]; !reflect.DeepEqual(got, []string{"tag-1", "tag-2"}) {
			t.Errorf("expected ids [tag-1 tag-2], got %v", got)
		}

		resp := ListTagsResponse{
			Tags: []Tag{{ID: "tag-1", Name: "production", ProjectID: "proj-123", UserID: "user-1"}},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListTags(context.Background(), &ListTagsOptions{
		Limit:         10,
		StartingAfter: "cursor-next",
		EndingBefore:  "cursor-prev",
		IDs:           []string{"tag-1", "tag-2"},
		OrgName:       "test-org",
		ProjectID:     "proj-123",
		ProjectName:   "example-project",
		TagName:       "production",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(result.Tags))
	}
	if result.Tags[0].ID != "tag-1" {
		t.Errorf("expected tag id tag-1, got %s", result.Tags[0].ID)
	}
}

func TestListTags_SpecialCharacters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/project_tag?project_name=Project+%26+Co&project_tag_name=v1%2Fbeta"
		if got := r.URL.RequestURI(); got != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, got)
		}

		resp := ListTagsResponse{Tags: []Tag{{ID: "tag-1"}}}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.ListTags(context.Background(), &ListTagsOptions{
		ProjectName: "Project & Co",
		TagName:     "v1/beta",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTag(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/project_tag" {
			t.Errorf("expected path /v1/project_tag, got %s", r.URL.Path)
		}

		var req CreateTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.ProjectID != "proj-123" {
			t.Errorf("expected project_id proj-123, got %q", req.ProjectID)
		}
		if req.Name != "production" {
			t.Errorf("expected name production, got %q", req.Name)
		}
		if req.Description != "Production workloads" {
			t.Errorf("expected description Production workloads, got %q", req.Description)
		}
		if req.Color != "#00AAFF" {
			t.Errorf("expected color #00AAFF, got %q", req.Color)
		}

		resp := Tag{
			ID:          "tag-123",
			Name:        req.Name,
			ProjectID:   req.ProjectID,
			Description: req.Description,
			Color:       req.Color,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	tag, err := client.CreateTag(context.Background(), &CreateTagRequest{
		ProjectID:   "proj-123",
		Name:        "production",
		Description: "Production workloads",
		Color:       "#00AAFF",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.ID != "tag-123" {
		t.Fatalf("expected id tag-123, got %q", tag.ID)
	}
}

func TestUpdateTag(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/project_tag/tag-123" {
			t.Errorf("expected path /v1/project_tag/tag-123, got %s", r.URL.Path)
		}

		var req UpdateTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Name == nil || *req.Name != "production-updated" {
			t.Fatalf("expected name production-updated, got %#v", req.Name)
		}
		if req.Description == nil || *req.Description != "Updated description" {
			t.Fatalf("expected description Updated description, got %#v", req.Description)
		}
		if req.Color == nil || *req.Color != "#11BB22" {
			t.Fatalf("expected color #11BB22, got %#v", req.Color)
		}

		resp := Tag{
			ID:          "tag-123",
			Name:        *req.Name,
			ProjectID:   "proj-123",
			Description: *req.Description,
			Color:       *req.Color,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	name := "production-updated"
	description := "Updated description"
	color := "#11BB22"
	tag, err := client.UpdateTag(context.Background(), "tag-123", &UpdateTagRequest{
		Name:        &name,
		Description: &description,
		Color:       &color,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.Name != "production-updated" {
		t.Fatalf("expected name production-updated, got %q", tag.Name)
	}
}

func TestUpdateTag_WhitespaceID(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
		t.Fatalf("expected no API call for whitespace-only ID")
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.UpdateTag(context.Background(), " \t\r\n", &UpdateTagRequest{})
	if !errors.Is(err, ErrEmptyTagID) {
		t.Fatalf("expected ErrEmptyTagID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API call for whitespace-only ID, got %d request(s)", requestCount)
	}
}

func TestDeleteTag(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/project_tag/tag-123" {
			t.Errorf("expected path /v1/project_tag/tag-123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	if err := client.DeleteTag(context.Background(), "tag-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTag_WhitespaceID(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
		t.Fatalf("expected no API call for whitespace-only ID")
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteTag(context.Background(), " \t\r\n")
	if !errors.Is(err, ErrEmptyTagID) {
		t.Fatalf("expected ErrEmptyTagID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API call for whitespace-only ID, got %d request(s)", requestCount)
	}
}
