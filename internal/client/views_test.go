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

func TestGetView_WhitespaceID(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
		t.Fatalf("expected no API call for whitespace-only ID")
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetView(context.Background(), "  \t \n", &GetViewOptions{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if !errors.Is(err, ErrEmptyViewID) {
		t.Fatalf("expected ErrEmptyViewID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API call for whitespace-only ID, got %d request(s)", requestCount)
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

func TestCreateView(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/view" {
			t.Errorf("expected path /v1/view, got %s", r.URL.Path)
		}

		var req CreateViewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.ObjectID != "project-123" {
			t.Errorf("expected object_id project-123, got %q", req.ObjectID)
		}
		if req.ObjectType != ACLObjectTypeProject {
			t.Errorf("expected object_type project, got %q", req.ObjectType)
		}
		if req.ViewType != ViewTypeExperiments {
			t.Errorf("expected view_type experiments, got %q", req.ViewType)
		}
		if req.Name != "default" {
			t.Errorf("expected name default, got %q", req.Name)
		}
		if got := req.Options["viewType"]; got != "table" {
			t.Errorf("expected options.viewType table, got %v", got)
		}
		if got := req.ViewData["search"]; got == nil {
			t.Errorf("expected view_data.search to be set")
		}

		resp := View{
			ID:         "view-123",
			Name:       req.Name,
			ObjectID:   req.ObjectID,
			ObjectType: req.ObjectType,
			ViewType:   req.ViewType,
			Options:    req.Options,
			ViewData:   req.ViewData,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	view, err := client.CreateView(context.Background(), &CreateViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
		ViewType:   ViewTypeExperiments,
		Name:       "default",
		Options: map[string]interface{}{
			"viewType": "table",
		},
		ViewData: map[string]interface{}{
			"search": map[string]interface{}{
				"filter": []interface{}{},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if view.ID != "view-123" {
		t.Fatalf("expected id view-123, got %s", view.ID)
	}
}

func TestUpdateView(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/view/view-123" {
			t.Errorf("expected path /v1/view/view-123, got %s", r.URL.Path)
		}

		var req map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		var objectID string
		if err := json.Unmarshal(req["object_id"], &objectID); err != nil {
			t.Fatalf("decode object_id: %v", err)
		}
		if objectID != "project-123" {
			t.Errorf("expected object_id project-123, got %q", objectID)
		}

		var objectType ACLObjectType
		if err := json.Unmarshal(req["object_type"], &objectType); err != nil {
			t.Fatalf("decode object_type: %v", err)
		}
		if objectType != ACLObjectTypeProject {
			t.Errorf("expected object_type project, got %q", objectType)
		}

		var name string
		if err := json.Unmarshal(req["name"], &name); err != nil {
			t.Fatalf("decode name: %v", err)
		}
		if name != "updated" {
			t.Errorf("expected name updated, got %q", name)
		}

		var options map[string]interface{}
		if err := json.Unmarshal(req["options"], &options); err != nil {
			t.Fatalf("decode options: %v", err)
		}
		if got := options["freezeColumns"]; got != true {
			t.Errorf("expected options.freezeColumns true, got %v", got)
		}

		resp := View{
			ID:         "view-123",
			Name:       name,
			ObjectID:   objectID,
			ObjectType: objectType,
			ViewType:   ViewTypeExperiments,
			Options:    options,
			ViewData: map[string]interface{}{
				"search": map[string]interface{}{
					"match": []interface{}{},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	name := "updated"
	options, err := viewJSONObjectRawMessage(map[string]interface{}{
		"freezeColumns": true,
	})
	if err != nil {
		t.Fatalf("encode options: %v", err)
	}
	viewData, err := viewJSONObjectRawMessage(map[string]interface{}{
		"search": map[string]interface{}{
			"match": []interface{}{},
		},
	})
	if err != nil {
		t.Fatalf("encode view_data: %v", err)
	}

	view, err := client.UpdateView(context.Background(), "view-123", &UpdateViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
		Name:       &name,
		Options:    options,
		ViewData:   viewData,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if view.Name != "updated" {
		t.Fatalf("expected updated name, got %q", view.Name)
	}
}

func TestUpdateView_SendsExplicitNullClears(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/view/view-123" {
			t.Errorf("expected path /v1/view/view-123, got %s", r.URL.Path)
		}

		var req map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if got := string(req["options"]); got != "null" {
			t.Fatalf("expected options to be explicit null, got %s", got)
		}
		if got := string(req["view_data"]); got != "null" {
			t.Fatalf("expected view_data to be explicit null, got %s", got)
		}
		if _, ok := req["name"]; ok {
			t.Fatalf("expected name to be omitted, got payload %v", req)
		}

		resp := View{
			ID:         "view-123",
			Name:       "default",
			ObjectID:   "project-123",
			ObjectType: ACLObjectTypeProject,
			ViewType:   ViewTypeExperiments,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.UpdateView(context.Background(), "view-123", &UpdateViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
		Options:    viewJSONNull(),
		ViewData:   viewJSONNull(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateView_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.UpdateView(context.Background(), "", &UpdateViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if !errors.Is(err, ErrEmptyViewID) {
		t.Fatalf("expected ErrEmptyViewID, got %v", err)
	}
}

func TestUpdateView_WhitespaceID(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
		t.Fatalf("expected no API call for whitespace-only ID")
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.UpdateView(context.Background(), " \t\r\n", &UpdateViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if !errors.Is(err, ErrEmptyViewID) {
		t.Fatalf("expected ErrEmptyViewID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API call for whitespace-only ID, got %d request(s)", requestCount)
	}
}

func TestDeleteView(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/view/view-123" {
			t.Errorf("expected path /v1/view/view-123, got %s", r.URL.Path)
		}

		var req DeleteViewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.ObjectID != "project-123" {
			t.Errorf("expected object_id project-123, got %q", req.ObjectID)
		}
		if req.ObjectType != ACLObjectTypeProject {
			t.Errorf("expected object_type project, got %q", req.ObjectType)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(View{ID: "view-123"})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteView(context.Background(), "view-123", &DeleteViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteView_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	err := client.DeleteView(context.Background(), "", &DeleteViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if !errors.Is(err, ErrEmptyViewID) {
		t.Fatalf("expected ErrEmptyViewID, got %v", err)
	}
}

func TestDeleteView_WhitespaceID(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
		t.Fatalf("expected no API call for whitespace-only ID")
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	err := client.DeleteView(context.Background(), " \t\r\n", &DeleteViewRequest{
		ObjectID:   "project-123",
		ObjectType: ACLObjectTypeProject,
	})
	if !errors.Is(err, ErrEmptyViewID) {
		t.Fatalf("expected ErrEmptyViewID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API call for whitespace-only ID, got %d request(s)", requestCount)
	}
}
