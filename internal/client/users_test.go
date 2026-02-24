package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestGetUser(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/user/user-123" {
			t.Errorf("expected path /v1/user/user-123, got %s", r.URL.Path)
		}

		resp := User{
			ID:         "user-123",
			GivenName:  "Jane",
			FamilyName: "Doe",
			Email:      "jane@example.com",
			AvatarURL:  "https://example.com/avatar.png",
			Created:    time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	user, err := client.GetUser(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-123" {
		t.Errorf("expected id user-123, got %s", user.ID)
	}
	if user.Email != "jane@example.com" {
		t.Errorf("expected email jane@example.com, got %s", user.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "User not found",
		})
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.GetUser(context.Background(), "missing-user")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr := &APIError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestListUsers_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/user" {
			t.Errorf("expected path /v1/user, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("limit"); got != "10" {
			t.Errorf("expected limit 10, got %q", got)
		}
		if got := query.Get("starting_after"); got != "cursor-next" {
			t.Errorf("expected starting_after cursor-next, got %q", got)
		}
		if got := query.Get("ending_before"); got != "cursor-prev" {
			t.Errorf("expected ending_before cursor-prev, got %q", got)
		}
		if got := query.Get("org_name"); got != "test-org" {
			t.Errorf("expected org_name test-org, got %q", got)
		}
		if got := query["ids"]; !reflect.DeepEqual(got, []string{"user-1", "user-2"}) {
			t.Errorf("expected ids [user-1 user-2], got %v", got)
		}
		if got := query["given_name"]; !reflect.DeepEqual(got, []string{"Jane", "John"}) {
			t.Errorf("expected given_name [Jane John], got %v", got)
		}
		if got := query["family_name"]; !reflect.DeepEqual(got, []string{"Doe"}) {
			t.Errorf("expected family_name [Doe], got %v", got)
		}
		if got := query["email"]; !reflect.DeepEqual(got, []string{"jane@example.com", "john@example.com"}) {
			t.Errorf("expected email [jane@example.com john@example.com], got %v", got)
		}

		resp := ListUsersResponse{
			Users: []User{{ID: "user-1", Email: "jane@example.com"}},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListUsers(context.Background(), &ListUsersOptions{
		Limit:         10,
		StartingAfter: "cursor-next",
		EndingBefore:  "cursor-prev",
		IDs:           []string{"user-1", "user-2"},
		GivenNames:    []string{"Jane", "John"},
		FamilyNames:   []string{"Doe"},
		Emails:        []string{"jane@example.com", "john@example.com"},
		OrgName:       "test-org",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(result.Users))
	}
	if result.Users[0].ID != "user-1" {
		t.Errorf("expected user id user-1, got %s", result.Users[0].ID)
	}
}

func TestListUsers_SpecialCharacters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/user?email=first.last%2Bops%40example.com&given_name=A%2FB&org_name=Acme+%26+Co"
		if got := r.URL.RequestURI(); got != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, got)
		}

		resp := ListUsersResponse{Users: []User{{ID: "user-1"}}}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.ListUsers(context.Background(), &ListUsersOptions{
		GivenNames: []string{"A/B"},
		Emails:     []string{"first.last+ops@example.com"},
		OrgName:    "Acme & Co",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUser_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetUser(context.Background(), "")
	if !errors.Is(err, ErrEmptyUserID) {
		t.Fatalf("expected ErrEmptyUserID, got %v", err)
	}
}
