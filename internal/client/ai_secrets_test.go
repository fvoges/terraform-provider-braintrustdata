package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestGetAISecret(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/ai_secret/ai-secret-123" {
			t.Errorf("expected path /v1/ai_secret/ai-secret-123, got %s", r.URL.Path)
		}

		resp := AISecret{
			ID:            "ai-secret-123",
			OrgID:         "org-123",
			Name:          "PROVIDER_OPENAI_CREDENTIAL",
			Type:          "openai",
			Metadata:      map[string]interface{}{"provider": "openai"},
			PreviewSecret: "sk-***1234",
			Created:       "2026-02-26T00:00:00Z",
			UpdatedAt:     "2026-02-26T01:00:00Z",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	aiSecret, err := client.GetAISecret(context.Background(), "ai-secret-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if aiSecret.ID != "ai-secret-123" {
		t.Fatalf("expected id ai-secret-123, got %s", aiSecret.ID)
	}
	if aiSecret.Name != "PROVIDER_OPENAI_CREDENTIAL" {
		t.Fatalf("expected name PROVIDER_OPENAI_CREDENTIAL, got %s", aiSecret.Name)
	}
}

func TestGetAISecret_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetAISecret(context.Background(), "")
	if !errors.Is(err, ErrEmptyAISecretID) {
		t.Fatalf("expected ErrEmptyAISecretID, got %v", err)
	}
}

func TestGetAISecret_WhitespaceID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetAISecret(context.Background(), "   ")
	if !errors.Is(err, ErrEmptyAISecretID) {
		t.Fatalf("expected ErrEmptyAISecretID, got %v", err)
	}
}

func TestListAISecrets_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/ai_secret" {
			t.Errorf("expected path /v1/ai_secret, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("org_name") != "test-org" {
			t.Errorf("expected org_name test-org, got %q", query.Get("org_name"))
		}
		if query.Get("ai_secret_name") != "PROVIDER_OPENAI_CREDENTIAL" {
			t.Errorf("expected ai_secret_name PROVIDER_OPENAI_CREDENTIAL, got %q", query.Get("ai_secret_name"))
		}
		if query.Get("limit") != "25" {
			t.Errorf("expected limit 25, got %q", query.Get("limit"))
		}
		if query.Get("starting_after") != "ai-secret-100" {
			t.Errorf("expected starting_after ai-secret-100, got %q", query.Get("starting_after"))
		}

		assertSameElements(t, query["ids"], []string{"ai-secret-1", "ai-secret-2"})
		assertSameElements(t, query["ai_secret_type"], []string{"openai", "anthropic"})

		resp := ListAISecretsResponse{
			AISecrets: []AISecret{
				{
					ID:            "ai-secret-101",
					OrgID:         "org-123",
					Name:          "PROVIDER_OPENAI_CREDENTIAL",
					Type:          "openai",
					PreviewSecret: "sk-***5678",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListAISecrets(context.Background(), &ListAISecretsOptions{
		OrgName:       "test-org",
		AISecretName:  "PROVIDER_OPENAI_CREDENTIAL",
		AISecretTypes: []string{"openai", "anthropic"},
		IDs:           []string{"ai-secret-1", "ai-secret-2"},
		Limit:         25,
		StartingAfter: "ai-secret-100",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.AISecrets) != 1 {
		t.Fatalf("expected 1 ai secret, got %d", len(result.AISecrets))
	}
	if result.AISecrets[0].ID != "ai-secret-101" {
		t.Fatalf("expected first ai secret id ai-secret-101, got %s", result.AISecrets[0].ID)
	}
}

func assertSameElements(t *testing.T, got, want []string) {
	t.Helper()

	gotCopy := append([]string(nil), got...)
	wantCopy := append([]string(nil), want...)
	slices.Sort(gotCopy)
	slices.Sort(wantCopy)

	if len(gotCopy) != len(wantCopy) {
		t.Fatalf("slice length mismatch: got=%v want=%v", gotCopy, wantCopy)
	}
	for i := range gotCopy {
		if gotCopy[i] != wantCopy[i] {
			t.Fatalf("slice content mismatch: got=%v want=%v", gotCopy, wantCopy)
		}
	}
}
