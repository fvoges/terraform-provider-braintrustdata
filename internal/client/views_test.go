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

func TestGetView(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/view/view-123" {
			t.Errorf("expected path /v1/view/view-123, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("object_id"); got != "project-123" {
			t.Errorf("expected object_id project-123, got %q", got)
		}
		if got := query.Get("object_type"); got != "project" {
			t.Errorf("expected object_type project, got %q", got)
		}

		resp := View{
			ID:         "view-123",
			Name:       "default",
			ObjectID:   "project-123",
			ObjectType: ACLObjectTypeProject,
			ViewType:   ViewTypeProjects,
			Created:    "2026-02-27T00:00:00Z",
			UserID:     "user-123",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	view, err := client.GetView(context.Background(), "view-123", &GetViewOptions{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if view.ID != "view-123" {
		t.Errorf("expected id view-123, got %s", view.ID)
	}
	if view.Name != "default" {
		t.Errorf("expected name default, got %s", view.Name)
	}
}

func TestGetView_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetView(context.Background(), "", &GetViewOptions{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if !errors.Is(err, ErrEmptyViewID) {
		t.Fatalf("expected ErrEmptyViewID, got %v", err)
	}
}

func TestListViews_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/view" {
			t.Errorf("expected path /v1/view, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("object_id"); got != "project-123" {
			t.Errorf("expected object_id project-123, got %q", got)
		}
		if got := query.Get("object_type"); got != "project" {
			t.Errorf("expected object_type project, got %q", got)
		}
		if got := query.Get("view_name"); got != "default" {
			t.Errorf("expected view_name default, got %q", got)
		}
		if got := query.Get("view_type"); got != "projects" {
			t.Errorf("expected view_type projects, got %q", got)
		}
		if got := query.Get("limit"); got != "10" {
			t.Errorf("expected limit 10, got %q", got)
		}
		if got := query.Get("starting_after"); got != "view-1" {
			t.Errorf("expected starting_after view-1, got %q", got)
		}
		if got := query.Get("ending_before"); got != "view-2" {
			t.Errorf("expected ending_before view-2, got %q", got)
		}
		if got := query["ids"]; !reflect.DeepEqual(got, []string{"view-a", "view-b"}) {
			t.Errorf("expected ids [view-a view-b], got %v", got)
		}

		resp := ListViewsResponse{
			Objects: []View{
				{
					ID:         "view-a",
					Name:       "default",
					ObjectID:   "project-123",
					ObjectType: ACLObjectTypeProject,
					ViewType:   ViewTypeProjects,
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListViews(context.Background(), &ListViewsOptions{
		ObjectID:      "project-123",
		ObjectType:    ACLObjectTypeProject,
		ViewName:      "default",
		ViewType:      ViewTypeProjects,
		Limit:         10,
		StartingAfter: "view-1",
		EndingBefore:  "view-2",
		IDs:           []string{"view-a", "view-b"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Objects) != 1 {
		t.Fatalf("expected 1 view, got %d", len(result.Objects))
	}
	if result.Objects[0].ID != "view-a" {
		t.Errorf("expected view id view-a, got %s", result.Objects[0].ID)
	}
}

func TestListViews_SpecialCharacters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/view?object_id=project%2F123&object_type=project&view_name=View+%26+QA"
		if got := r.URL.RequestURI(); got != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, got)
		}

		resp := ListViewsResponse{Objects: []View{{ID: "view-1"}}}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.ListViews(context.Background(), &ListViewsOptions{
		ObjectID:   "project/123",
		ObjectType: ACLObjectTypeProject,
		ViewName:   "View & QA",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
