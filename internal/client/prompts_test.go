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
