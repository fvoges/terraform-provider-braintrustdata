package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEnvironmentVariable(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/env_var/env-var-123" {
			t.Errorf("expected path /v1/env_var/env-var-123, got %s", r.URL.Path)
		}

		resp := EnvironmentVariable{
			ID:         "env-var-123",
			ObjectType: "project",
			ObjectID:   "project-123",
			Name:       "OPENAI_API_KEY",
			Created:    "2026-02-26T00:00:00Z",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	envVar, err := client.GetEnvironmentVariable(context.Background(), "env-var-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if envVar.ID != "env-var-123" {
		t.Fatalf("expected id env-var-123, got %s", envVar.ID)
	}
	if envVar.Name != "OPENAI_API_KEY" {
		t.Fatalf("expected name OPENAI_API_KEY, got %s", envVar.Name)
	}
}

func TestGetEnvironmentVariable_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetEnvironmentVariable(context.Background(), "")
	if !errors.Is(err, ErrEmptyEnvironmentVariableID) {
		t.Fatalf("expected ErrEmptyEnvironmentVariableID, got %v", err)
	}
}

func TestGetEnvironmentVariable_WhitespaceID_DoesNotCallAPI(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetEnvironmentVariable(context.Background(), " \t\n ")
	if !errors.Is(err, ErrEmptyEnvironmentVariableID) {
		t.Fatalf("expected ErrEmptyEnvironmentVariableID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API requests, got %d", requestCount)
	}
}

func TestListEnvironmentVariables_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		expectedRequestURI := "/v1/env_var?limit=25&object_id=project-123&object_type=project&starting_after=env-var-100"
		if got := r.URL.RequestURI(); got != expectedRequestURI {
			t.Errorf("expected request URI %q, got %q", expectedRequestURI, got)
		}

		resp := ListEnvironmentVariablesResponse{
			EnvironmentVariables: []EnvironmentVariable{
				{
					ID:         "env-var-101",
					ObjectType: "project",
					ObjectID:   "project-123",
					Name:       "OPENAI_API_KEY",
					Created:    "2026-02-26T00:00:00Z",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListEnvironmentVariables(context.Background(), &ListEnvironmentVariablesOptions{
		ObjectType:    "project",
		ObjectID:      "project-123",
		Limit:         25,
		StartingAfter: "env-var-100",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.EnvironmentVariables) != 1 {
		t.Fatalf("expected 1 environment variable, got %d", len(result.EnvironmentVariables))
	}
	if result.EnvironmentVariables[0].ID != "env-var-101" {
		t.Fatalf("expected first environment variable id env-var-101, got %s", result.EnvironmentVariables[0].ID)
	}
}

func TestListEnvironmentVariables_SpecialCharacters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedRequestURI := "/v1/env_var?object_id=proj%2Ftest+1&object_type=project"
		if got := r.URL.RequestURI(); got != expectedRequestURI {
			t.Errorf("expected request URI %q, got %q", expectedRequestURI, got)
		}

		resp := ListEnvironmentVariablesResponse{}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.ListEnvironmentVariables(context.Background(), &ListEnvironmentVariablesOptions{
		ObjectType: "project",
		ObjectID:   "proj/test 1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
