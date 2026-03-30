package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestCreateAISecret(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/ai_secret" {
			t.Errorf("expected path /v1/ai_secret, got %s", r.URL.Path)
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		if payload["name"] != "PROVIDER_OPENAI_CREDENTIAL" {
			t.Errorf("expected name PROVIDER_OPENAI_CREDENTIAL, got %v", payload["name"])
		}
		if payload["type"] != "openai" {
			t.Errorf("expected type openai, got %v", payload["type"])
		}
		if payload["secret"] != "sk-secret" {
			t.Errorf("expected secret sk-secret, got %v", payload["secret"])
		}
		if payload["org_name"] != "test-org" {
			t.Errorf("expected org_name test-org, got %v", payload["org_name"])
		}

		metadataRaw, ok := payload["metadata"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected metadata object, got %T", payload["metadata"])
		}
		if metadataRaw["provider"] != "openai" {
			t.Errorf("expected metadata.provider=openai, got %v", metadataRaw["provider"])
		}

		resp := AISecret{
			ID:            "ai-secret-321",
			Name:          "PROVIDER_OPENAI_CREDENTIAL",
			Type:          "openai",
			PreviewSecret: "********",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	created, err := client.CreateAISecret(context.Background(), &CreateAISecretRequest{
		Name:   "PROVIDER_OPENAI_CREDENTIAL",
		Type:   "openai",
		Secret: "sk-secret",
		Metadata: map[string]interface{}{
			"provider": "openai",
		},
		OrgName: "test-org",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.ID != "ai-secret-321" {
		t.Fatalf("expected created id ai-secret-321, got %s", created.ID)
	}
	if created.Name != "PROVIDER_OPENAI_CREDENTIAL" {
		t.Fatalf("expected created name PROVIDER_OPENAI_CREDENTIAL, got %s", created.Name)
	}
}

func TestCreateAISecret_NilRequest(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.CreateAISecret(context.Background(), nil)
	if !errors.Is(err, ErrNilCreateAISecretRequest) {
		t.Fatalf("expected ErrNilCreateAISecretRequest, got %v", err)
	}
}

func TestCreateAISecretRequestSecretJSONTag(t *testing.T) {
	t.Parallel()

	field, ok := reflect.TypeOf(CreateAISecretRequest{}).FieldByName("Secret")
	if !ok {
		t.Fatal("expected Secret field on CreateAISecretRequest")
	}

	if got := field.Tag.Get("json"); got != "secret" {
		t.Fatalf("expected Secret json tag to be %q, got %q", "secret", got)
	}
}

func TestUpdateAISecret(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/ai_secret/ai-secret-123" {
			t.Errorf("expected path /v1/ai_secret/ai-secret-123, got %s", r.URL.Path)
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		if payload["name"] != "PROVIDER_OPENAI_CREDENTIAL" {
			t.Errorf("expected name PROVIDER_OPENAI_CREDENTIAL, got %v", payload["name"])
		}
		if payload["type"] != "anthropic" {
			t.Errorf("expected type anthropic, got %v", payload["type"])
		}
		if payload["secret"] != "new-secret" {
			t.Errorf("expected secret new-secret, got %v", payload["secret"])
		}

		if _, ok := payload["org_name"]; ok {
			t.Errorf("expected org_name to be omitted on update payload, got %v", payload["org_name"])
		}

		resp := AISecret{
			ID:            "ai-secret-123",
			Name:          "PROVIDER_OPENAI_CREDENTIAL",
			Type:          "anthropic",
			PreviewSecret: "********",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	name := "PROVIDER_OPENAI_CREDENTIAL"
	secretType := "anthropic"
	secretValue := "new-secret"
	updated, err := client.UpdateAISecret(context.Background(), "ai-secret-123", &UpdateAISecretRequest{
		Name:   &name,
		Type:   &secretType,
		Secret: &secretValue,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Type != "anthropic" {
		t.Fatalf("expected updated type anthropic, got %s", updated.Type)
	}
}

func TestUpdateAISecret_NilRequest(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.UpdateAISecret(context.Background(), "ai-secret-123", nil)
	if !errors.Is(err, ErrNilUpdateAISecretRequest) {
		t.Fatalf("expected ErrNilUpdateAISecretRequest, got %v", err)
	}
}

func TestUpdateAISecret_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.UpdateAISecret(context.Background(), " ", &UpdateAISecretRequest{})
	if !errors.Is(err, ErrEmptyAISecretID) {
		t.Fatalf("expected ErrEmptyAISecretID, got %v", err)
	}
}

func TestDeleteAISecret(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/ai_secret/ai-secret-123" {
			t.Errorf("expected path /v1/ai_secret/ai-secret-123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	if err := client.DeleteAISecret(context.Background(), "ai-secret-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAISecret_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	err := client.DeleteAISecret(context.Background(), "")
	if !errors.Is(err, ErrEmptyAISecretID) {
		t.Fatalf("expected ErrEmptyAISecretID, got %v", err)
	}
}
