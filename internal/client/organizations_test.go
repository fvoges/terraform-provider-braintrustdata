package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestGetOrganization(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organization/org-123" {
			t.Errorf("expected path /v1/organization/org-123, got %s", r.URL.Path)
		}

		_ = json.NewEncoder(w).Encode(Organization{
			ID:                 "org-123",
			Name:               "Acme",
			APIURL:             ptrString("https://api.acme.dev"),
			IsUniversalAPI:     ptrBool(true),
			IsDataplanePrivate: ptrBool(false),
			ProxyURL:           ptrString("https://proxy.acme.dev"),
			RealtimeURL:        ptrString("wss://realtime.acme.dev"),
			Created:            ptrString("2026-02-26T00:00:00Z"),
			ImageRenderingMode: ptrString("auto"),
		})
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-default")
	c.httpClient = server.Client()

	org, err := c.GetOrganization(context.Background(), "org-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if org.ID != "org-123" {
		t.Fatalf("id mismatch: got=%q", org.ID)
	}
	if org.Name != "Acme" {
		t.Fatalf("name mismatch: got=%q", org.Name)
	}
	if org.APIURL == nil || *org.APIURL != "https://api.acme.dev" {
		t.Fatalf("api_url mismatch: got=%v", org.APIURL)
	}
	if org.IsUniversalAPI == nil || !*org.IsUniversalAPI {
		t.Fatalf("is_universal_api mismatch: got=%v", org.IsUniversalAPI)
	}
	if org.IsDataplanePrivate == nil || *org.IsDataplanePrivate {
		t.Fatalf("is_dataplane_private mismatch: got=%v", org.IsDataplanePrivate)
	}
}

func TestUpdateOrganization(t *testing.T) {
	t.Parallel()

	type capturedRequest struct {
		Name               *string `json:"name,omitempty"`
		IsUniversalAPI     *bool   `json:"is_universal_api,omitempty"`
		IsDataplanePrivate *bool   `json:"is_dataplane_private,omitempty"`
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organization/org-123" {
			t.Errorf("expected path /v1/organization/org-123, got %s", r.URL.Path)
		}

		var req capturedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Name == nil || *req.Name != "Acme Updated" {
			t.Fatalf("expected name to be sent, got %v", req.Name)
		}
		if req.IsUniversalAPI == nil || *req.IsUniversalAPI {
			t.Fatalf("expected is_universal_api=false to be sent, got %v", req.IsUniversalAPI)
		}
		if req.IsDataplanePrivate == nil || !*req.IsDataplanePrivate {
			t.Fatalf("expected is_dataplane_private=true to be sent, got %v", req.IsDataplanePrivate)
		}

		_ = json.NewEncoder(w).Encode(Organization{
			ID:                 "org-123",
			Name:               "Acme Updated",
			IsUniversalAPI:     ptrBool(false),
			IsDataplanePrivate: ptrBool(true),
		})
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-default")
	c.httpClient = server.Client()

	org, err := c.UpdateOrganization(context.Background(), "org-123", &PatchOrganizationRequest{
		Name:               ptrString("Acme Updated"),
		IsUniversalAPI:     ptrBool(false),
		IsDataplanePrivate: ptrBool(true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if org.Name != "Acme Updated" {
		t.Fatalf("name mismatch: got=%q", org.Name)
	}
	if org.IsUniversalAPI == nil || *org.IsUniversalAPI {
		t.Fatalf("is_universal_api mismatch: got=%v", org.IsUniversalAPI)
	}
	if org.IsDataplanePrivate == nil || !*org.IsDataplanePrivate {
		t.Fatalf("is_dataplane_private mismatch: got=%v", org.IsDataplanePrivate)
	}
}

func TestGetOrganization_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Organization not found"})
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-default")
	c.httpClient = server.Client()

	_, err := c.GetOrganization(context.Background(), "org-missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr := &APIError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetOrganization_SpecialCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		orgID string
	}{
		{
			name:  "slash",
			orgID: "org/123",
		},
		{
			name:  "space",
			orgID: "org 123",
		},
		{
			name:  "unicode",
			orgID: "組織",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expectedEscapedPath := "/v1/organization/" + url.PathEscape(tc.orgID)
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.EscapedPath() != expectedEscapedPath {
					t.Fatalf("escaped path mismatch: got=%q want=%q", r.URL.EscapedPath(), expectedEscapedPath)
				}
				_ = json.NewEncoder(w).Encode(Organization{ID: tc.orgID, Name: "Name"})
			}))
			defer server.Close()

			c := NewClient("sk-test", server.URL, "org-default")
			c.httpClient = server.Client()

			org, err := c.GetOrganization(context.Background(), tc.orgID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(org, &Organization{ID: tc.orgID, Name: "Name"}) {
				t.Fatalf("unexpected organization: %#v", org)
			}
		})
	}
}

func TestOrganization_EmptyID(t *testing.T) {
	t.Parallel()

	c := NewClient("sk-test", "https://api.example.com", "org-default")

	_, err := c.GetOrganization(context.Background(), "")
	if !errors.Is(err, ErrEmptyOrganizationID) {
		t.Fatalf("expected ErrEmptyOrganizationID from GetOrganization, got %v", err)
	}

	_, err = c.UpdateOrganization(context.Background(), "", &PatchOrganizationRequest{Name: ptrString("x")})
	if !errors.Is(err, ErrEmptyOrganizationID) {
		t.Fatalf("expected ErrEmptyOrganizationID from UpdateOrganization, got %v", err)
	}
}

func ptrString(v string) *string {
	return &v
}

func ptrBool(v bool) *bool {
	return &v
}
