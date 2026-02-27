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
