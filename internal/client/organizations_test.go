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

func TestListOrganizations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		options      *ListOrganizationsOptions
		name         string
		expectedPath string
	}{
		{
			name:         "without options",
			options:      nil,
			expectedPath: "/v1/organization",
		},
		{
			name: "with org name filter",
			options: &ListOrganizationsOptions{
				OrgName: "acme",
			},
			expectedPath: "/v1/organization?org_name=acme",
		},
		{
			name: "with pagination and limit",
			options: &ListOrganizationsOptions{
				OrgName:       "acme",
				StartingAfter: "org-1",
				Limit:         10,
			},
			expectedPath: "/v1/organization?limit=10&org_name=acme&starting_after=org-1",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath := r.URL.Path
				if r.URL.RawQuery != "" {
					gotPath += "?" + r.URL.RawQuery
				}
				if gotPath != tc.expectedPath {
					t.Fatalf("request path mismatch: got=%q want=%q", gotPath, tc.expectedPath)
				}

				_ = json.NewEncoder(w).Encode(ListOrganizationsResponse{
					Organizations: []Organization{
						{ID: "org-1", Name: "Acme"},
						{ID: "org-2", Name: "Beta"},
					},
				})
			}))
			defer server.Close()

			c := NewClient("sk-test", server.URL, "org-default")
			c.httpClient = server.Client()

			resp, err := c.ListOrganizations(context.Background(), tc.options)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.Organizations) != 2 {
				t.Fatalf("organization count mismatch: got=%d", len(resp.Organizations))
			}
			if resp.Organizations[0].ID != "org-1" {
				t.Fatalf("first organization mismatch: %#v", resp.Organizations[0])
			}
		})
	}
}

func TestListOrganizations_SpecialCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		options       *ListOrganizationsOptions
		expectedQuery string
	}{
		{
			name: "org name with spaces and symbols",
			options: &ListOrganizationsOptions{
				OrgName: "Acme & Co + Labs",
			},
			expectedQuery: "org_name=Acme+%26+Co+%2B+Labs",
		},
		{
			name: "cursor with slash and unicode",
			options: &ListOrganizationsOptions{
				StartingAfter: "org/組織",
			},
			expectedQuery: "starting_after=org%2F%E7%B5%84%E7%B9%94",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RawQuery != tc.expectedQuery {
					t.Fatalf("query mismatch: got=%q want=%q", r.URL.RawQuery, tc.expectedQuery)
				}
				_ = json.NewEncoder(w).Encode(ListOrganizationsResponse{})
			}))
			defer server.Close()

			c := NewClient("sk-test", server.URL, "org-default")
			c.httpClient = server.Client()

			_, err := c.ListOrganizations(context.Background(), tc.options)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestListOrganizations_ErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-default")
	c.httpClient = server.Client()

	_, err := c.ListOrganizations(context.Background(), &ListOrganizationsOptions{OrgName: "acme"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr := &APIError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Message == "" {
		t.Fatalf("expected error message, got empty: %+v", apiErr)
	}
}

func TestListOrganizations_QueryParameterOrder(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotValues, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			t.Fatalf("failed to parse query: %v", err)
		}
		expectedValues, err := url.ParseQuery("ending_before=org-0&limit=5&org_name=acme&starting_after=org-1")
		if err != nil {
			t.Fatalf("failed to parse expected query: %v", err)
		}
		if !reflect.DeepEqual(gotValues, expectedValues) {
			t.Fatalf("query mismatch: got=%v want=%v", gotValues, expectedValues)
		}
		_ = json.NewEncoder(w).Encode(ListOrganizationsResponse{})
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-default")
	c.httpClient = server.Client()

	_, err := c.ListOrganizations(context.Background(), &ListOrganizationsOptions{
		OrgName:       "acme",
		StartingAfter: "org-1",
		EndingBefore:  "org-0",
		Limit:         5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ptrString(v string) *string {
	return &v
}

func ptrBool(v bool) *bool {
	return &v
}
