package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnvironmentVariableUsedUnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input   string
		wantErr string
		want    bool
	}{
		"bool true": {
			input: `{"used":true}`,
			want:  true,
		},
		"string true": {
			input: `{"used":"true"}`,
			want:  true,
		},
		"string false": {
			input: `{"used":"false"}`,
			want:  false,
		},
		"timestamp string": {
			input: `{"used":"2026-03-11T12:34:56Z"}`,
			want:  true,
		},
		"empty string": {
			input: `{"used":""}`,
			want:  false,
		},
		"null": {
			input: `{"used":null}`,
			want:  false,
		},
		"number": {
			input:   `{"used":123}`,
			wantErr: "unsupported used value: 123",
		},
		"object": {
			input:   `{"used":{"at":"2026-03-11T12:34:56Z"}}`,
			wantErr: `unsupported used value: {"at":"2026-03-11T12:34:56Z"}`,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var envVar EnvironmentVariable
			err := json.Unmarshal([]byte(tc.input), &envVar)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("error mismatch: got=%q want=%q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := bool(envVar.Used); got != tc.want {
				t.Fatalf("used mismatch: got=%t want=%t", got, tc.want)
			}
		})
	}
}

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
			Created:    "2024-01-15T10:30:00Z",
			Used:       true,
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
	if got := bool(envVar.Used); !got {
		t.Fatal("expected used=true")
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
					Created:    "2024-01-15T10:30:00Z",
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

func TestCreateEnvironmentVariable(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/env_var" {
			t.Errorf("expected path /v1/env_var, got %s", r.URL.Path)
		}

		var req map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		for _, key := range []string{"name", "object_type", "object_id", "value"} {
			if _, ok := req[key]; !ok {
				t.Errorf("expected %q in create payload, got %v", key, req)
			}
		}
		if _, ok := req["description"]; ok {
			t.Errorf("expected description to be omitted from create payload, got %v", req)
		}

		resp := EnvironmentVariable{
			ID:             "env-var-123",
			ObjectType:     "project",
			ObjectID:       "project-123",
			Name:           "OPENAI_API_KEY",
			Description:    "OpenAI API key",
			SecretType:     "text",
			SecretCategory: "api",
			Created:        "2024-01-15T10:30:00Z",
			Used:           false,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	envVar, err := client.CreateEnvironmentVariable(context.Background(), &CreateEnvironmentVariableRequest{
		ObjectType: "project",
		ObjectID:   "project-123",
		Name:       "OPENAI_API_KEY",
		Value:      "sk-test-value",
		Metadata: map[string]interface{}{
			"owner": "ml-platform",
		},
		SecretType:     "text",
		SecretCategory: "api",
	})
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

func TestUpdateEnvironmentVariable_SendsExplicitClearFields(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/env_var/env-var-123" {
			t.Errorf("expected path /v1/env_var/env-var-123, got %s", r.URL.Path)
		}

		var req map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		for _, key := range []string{"name", "metadata", "value"} {
			if _, ok := req[key]; !ok {
				t.Errorf("expected %q in update payload, got %v", key, req)
			}
		}
		if _, ok := req["description"]; ok {
			t.Errorf("expected description to be omitted from update payload, got %v", req)
		}

		resp := EnvironmentVariable{
			ID:          "env-var-123",
			ObjectType:  "project",
			ObjectID:    "project-123",
			Name:        "OPENAI_API_KEY_V2",
			Description: "",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.UpdateEnvironmentVariable(context.Background(), "env-var-123", &UpdateEnvironmentVariableRequest{
		Name:     envVarStringPointer("OPENAI_API_KEY_V2"),
		Metadata: envVarMapPointer(map[string]interface{}{}),
		Value:    envVarStringPointer("new-value"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEnvironmentVariable_WhitespaceID_DoesNotCallAPI(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		requestCount++
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.UpdateEnvironmentVariable(context.Background(), " \t\n ", &UpdateEnvironmentVariableRequest{
		Name: envVarStringPointer("OPENAI_API_KEY"),
	})
	if !errors.Is(err, ErrEmptyEnvironmentVariableID) {
		t.Fatalf("expected ErrEmptyEnvironmentVariableID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API requests, got %d", requestCount)
	}
}

func TestDeleteEnvironmentVariable(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.RequestURI() != "/v1/env_var/env-var%2F123" {
			t.Errorf("expected escaped request URI /v1/env_var/env-var%%2F123, got %s", r.URL.RequestURI())
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	if err := client.DeleteEnvironmentVariable(context.Background(), "env-var/123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteEnvironmentVariable_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	err := client.DeleteEnvironmentVariable(context.Background(), "")
	if !errors.Is(err, ErrEmptyEnvironmentVariableID) {
		t.Fatalf("expected ErrEmptyEnvironmentVariableID, got %v", err)
	}
}

func envVarStringPointer(v string) *string {
	return &v
}

func envVarMapPointer(v map[string]interface{}) *map[string]interface{} {
	return &v
}
