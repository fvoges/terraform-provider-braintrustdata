package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPrompt(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/prompt/prompt-123" {
			t.Errorf("expected path /v1/prompt/prompt-123, got %s", r.URL.Path)
		}

		resp := Prompt{
			ID:          "prompt-123",
			ProjectID:   "project-123",
			Name:        "support-agent",
			Slug:        "support-agent",
			Description: "Support assistant prompt",
			Created:     "2026-02-27T00:00:00Z",
			UserID:      "user-123",
			OrgID:       "org-test",
			Metadata: map[string]interface{}{
				"owner": "ml-team",
			},
			Tags: []string{"production", "support"},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	prompt, err := client.GetPrompt(context.Background(), "prompt-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if prompt.ID != "prompt-123" {
		t.Errorf("expected id prompt-123, got %s", prompt.ID)
	}
	if prompt.ProjectID != "project-123" {
		t.Errorf("expected project_id project-123, got %s", prompt.ProjectID)
	}
	if prompt.Name != "support-agent" {
		t.Errorf("expected name support-agent, got %s", prompt.Name)
	}
	if prompt.Slug != "support-agent" {
		t.Errorf("expected slug support-agent, got %s", prompt.Slug)
	}
}

func TestGetPrompt_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetPrompt(context.Background(), "")
	if !errors.Is(err, ErrEmptyPromptID) {
		t.Fatalf("expected ErrEmptyPromptID, got %v", err)
	}
}

func TestGetPrompt_WhitespaceOnlyID(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetPrompt(context.Background(), "   \n\t ")
	if !errors.Is(err, ErrEmptyPromptID) {
		t.Fatalf("expected ErrEmptyPromptID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API call for whitespace-only ID, got %d request(s)", requestCount)
	}
}

func TestListPrompts_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/prompt" {
			t.Errorf("expected path /v1/prompt, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("project_id"); got != "project-123" {
			t.Errorf("expected project_id project-123, got %q", got)
		}
		if got := query.Get("prompt_name"); got != "support-agent" {
			t.Errorf("expected prompt_name support-agent, got %q", got)
		}
		if got := query.Get("slug"); got != "support-agent" {
			t.Errorf("expected slug support-agent, got %q", got)
		}
		if got := query.Get("version"); got != "v1" {
			t.Errorf("expected version v1, got %q", got)
		}
		if got := query.Get("limit"); got != "10" {
			t.Errorf("expected limit 10, got %q", got)
		}
		if got := query.Get("starting_after"); got != "prompt-1" {
			t.Errorf("expected starting_after prompt-1, got %q", got)
		}
		if got := query.Get("ending_before"); got != "prompt-2" {
			t.Errorf("expected ending_before prompt-2, got %q", got)
		}

		resp := ListPromptsResponse{
			Prompts: []Prompt{
				{ID: "prompt-1", ProjectID: "project-123", Name: "support-agent"},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListPrompts(context.Background(), &ListPromptsOptions{
		ProjectID:     "project-123",
		PromptName:    "support-agent",
		Slug:          "support-agent",
		Version:       "v1",
		Limit:         10,
		StartingAfter: "prompt-1",
		EndingBefore:  "prompt-2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(result.Prompts))
	}
	if result.Prompts[0].ID != "prompt-1" {
		t.Errorf("expected prompt id prompt-1, got %s", result.Prompts[0].ID)
	}
}

func TestListPrompts_SpecialCharacters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/prompt?project_id=project%2F123&prompt_name=Support+%26+QA"
		if got := r.URL.RequestURI(); got != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, got)
		}

		resp := ListPromptsResponse{Prompts: []Prompt{{ID: "prompt-1"}}}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.ListPrompts(context.Background(), &ListPromptsOptions{
		ProjectID:  "project/123",
		PromptName: "Support & QA",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePrompt(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/prompt" {
			t.Errorf("expected path /v1/prompt, got %s", r.URL.Path)
		}

		var req CreatePromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Name != "support-agent" {
			t.Errorf("expected name support-agent, got %s", req.Name)
		}
		if req.ProjectID != "project-123" {
			t.Errorf("expected project_id project-123, got %s", req.ProjectID)
		}

		resp := Prompt{
			ID:          "prompt-abc",
			ProjectID:   "project-123",
			Name:        "support-agent",
			Slug:        "support-agent",
			Description: "Support assistant",
			Created:     "2026-02-27T00:00:00Z",
			OrgID:       "org-test",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	prompt, err := c.CreatePrompt(context.Background(), &CreatePromptRequest{
		ProjectID:   "project-123",
		Name:        "support-agent",
		Description: "Support assistant",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt.ID != "prompt-abc" {
		t.Errorf("expected id prompt-abc, got %s", prompt.ID)
	}
	if prompt.Name != "support-agent" {
		t.Errorf("expected name support-agent, got %s", prompt.Name)
	}
}

func TestCreatePrompt_WithTagsAndMetadata(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreatePromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if len(req.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(req.Tags))
		}
		if req.Metadata["env"] != "prod" {
			t.Errorf("expected metadata env=prod, got %v", req.Metadata["env"])
		}

		resp := Prompt{
			ID:        "prompt-tagged",
			ProjectID: "project-123",
			Name:      "tagged-prompt",
			Tags:      req.Tags,
			Metadata:  req.Metadata,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	prompt, err := c.CreatePrompt(context.Background(), &CreatePromptRequest{
		ProjectID: "project-123",
		Name:      "tagged-prompt",
		Tags:      []string{"ml", "production"},
		Metadata:  map[string]interface{}{"env": "prod"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt.ID != "prompt-tagged" {
		t.Errorf("expected id prompt-tagged, got %s", prompt.ID)
	}
}

func TestUpdatePrompt(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/prompt/prompt-abc" {
			t.Errorf("expected path /v1/prompt/prompt-abc, got %s", r.URL.Path)
		}

		var req UpdatePromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		resp := Prompt{
			ID:          "prompt-abc",
			ProjectID:   "project-123",
			Name:        promptDerefString(req.Name),
			Description: promptDerefString(req.Description),
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	prompt, err := c.UpdatePrompt(context.Background(), "prompt-abc", &UpdatePromptRequest{
		Name:        stringPointer("support-agent-v2"),
		Description: stringPointer("Updated description"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt.Name != "support-agent-v2" {
		t.Errorf("expected name support-agent-v2, got %s", prompt.Name)
	}
}

func TestUpdatePrompt_SendsExplicitClearFields(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/prompt/prompt-abc" {
			t.Errorf("expected path /v1/prompt/prompt-abc, got %s", r.URL.Path)
		}

		var req map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		for _, key := range []string{"name", "description", "metadata", "tags", "prompt_data"} {
			if _, ok := req[key]; !ok {
				t.Errorf("expected %q to be present in update payload, body=%v", key, req)
			}
		}

		resp := Prompt{
			ID:        "prompt-abc",
			ProjectID: "project-123",
			Name:      "support-agent-v2",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	_, err := c.UpdatePrompt(context.Background(), "prompt-abc", &UpdatePromptRequest{
		Name:        stringPointer("support-agent-v2"),
		Description: stringPointer(""),
		Metadata:    mapPointer(map[string]interface{}{}),
		Tags:        stringSlicePointer([]string{}),
		PromptData:  interfacePointer(nil),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePrompt_EmptyID(t *testing.T) {
	c := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := c.UpdatePrompt(context.Background(), "", &UpdatePromptRequest{Name: stringPointer("test")})
	if !errors.Is(err, ErrEmptyPromptID) {
		t.Fatalf("expected ErrEmptyPromptID, got %v", err)
	}
}

func stringPointer(v string) *string {
	return &v
}

func stringSlicePointer(v []string) *[]string {
	return &v
}

func mapPointer(v map[string]interface{}) *map[string]interface{} {
	return &v
}

func interfacePointer(v interface{}) *interface{} {
	return &v
}

func promptDerefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func TestDeletePrompt(t *testing.T) {
	var deletedID string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		deletedID = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "prompt-abc"})
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	err := c.DeletePrompt(context.Background(), "prompt-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != "/v1/prompt/prompt-abc" {
		t.Errorf("expected DELETE /v1/prompt/prompt-abc, got %s", deletedID)
	}
}

func TestDeletePrompt_EmptyID(t *testing.T) {
	c := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	err := c.DeletePrompt(context.Background(), "")
	if !errors.Is(err, ErrEmptyPromptID) {
		t.Fatalf("expected ErrEmptyPromptID, got %v", err)
	}
}

func TestDeletePrompt_SpecialCharactersInID(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Go HTTP servers decode %2F in r.URL.Path; check RawPath for encoding.
		rawPath := r.URL.RawPath
		if rawPath == "" {
			rawPath = r.URL.Path
		}
		if rawPath != "/v1/prompt/prompt%2Fspecial" {
			t.Errorf("expected raw path /v1/prompt/prompt%%2Fspecial, got %s", rawPath)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "prompt/special"})
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	err := c.DeletePrompt(context.Background(), "prompt/special")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
