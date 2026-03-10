package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFunction(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/function/function-123" {
			t.Errorf("expected path /v1/function/function-123, got %s", r.URL.Path)
		}

		resp := Function{
			XactID:       "xact-1",
			Created:      "2026-03-10T00:00:00Z",
			Description:  "Tool function",
			FunctionData: map[string]interface{}{"runtime": "python"},
			FunctionSchema: map[string]interface{}{
				"type": "object",
			},
			FunctionType: "tool",
			ID:           "function-123",
			LogID:        "log-123",
			Metadata:     map[string]interface{}{"owner": "ml"},
			Name:         "my-tool",
			OrgID:        "org-123",
			Origin:       map[string]interface{}{"source": "api"},
			ProjectID:    "project-123",
			PromptData:   map[string]interface{}{"prompt": "Hello"},
			Slug:         "my-tool",
			Tags:         []string{"prod", "tool"},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	fn, err := c.GetFunction(context.Background(), "function-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn.ID != "function-123" {
		t.Fatalf("expected id function-123, got %q", fn.ID)
	}
	if fn.Name != "my-tool" {
		t.Fatalf("expected name my-tool, got %q", fn.Name)
	}
	if fn.FunctionType != "tool" {
		t.Fatalf("expected function_type tool, got %q", fn.FunctionType)
	}
}

func TestGetFunction_EmptyID(t *testing.T) {
	c := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := c.GetFunction(context.Background(), "")
	if !errors.Is(err, ErrEmptyFunctionID) {
		t.Fatalf("expected ErrEmptyFunctionID, got %v", err)
	}
}

func TestGetFunction_WhitespaceOnlyID(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	_, err := c.GetFunction(context.Background(), "  \n\t ")
	if !errors.Is(err, ErrEmptyFunctionID) {
		t.Fatalf("expected ErrEmptyFunctionID, got %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected no API call for whitespace-only ID, got %d request(s)", requestCount)
	}
}

func TestListFunctions_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/function" {
			t.Errorf("expected path /v1/function, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("project_id"); got != "project-123" {
			t.Errorf("expected project_id project-123, got %q", got)
		}
		if got := query.Get("function_name"); got != "tool-a" {
			t.Errorf("expected function_name tool-a, got %q", got)
		}
		if got := query.Get("slug"); got != "tool-a" {
			t.Errorf("expected slug tool-a, got %q", got)
		}
		if got := query.Get("limit"); got != "0" {
			t.Errorf("expected limit 0, got %q", got)
		}
		if got := query.Get("starting_after"); got != "function-1" {
			t.Errorf("expected starting_after function-1, got %q", got)
		}
		if got := query.Get("ending_before"); got != "function-2" {
			t.Errorf("expected ending_before function-2, got %q", got)
		}

		resp := ListFunctionsResponse{
			Functions: []Function{{ID: "function-1", Name: "tool-a"}},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	limit := 0
	result, err := c.ListFunctions(context.Background(), &ListFunctionsOptions{
		ProjectID:     "project-123",
		FunctionName:  "tool-a",
		Slug:          "tool-a",
		Limit:         &limit,
		StartingAfter: "function-1",
		EndingBefore:  "function-2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(result.Functions))
	}
	if result.Functions[0].ID != "function-1" {
		t.Fatalf("expected function id function-1, got %s", result.Functions[0].ID)
	}
}

func TestListFunctions_SpecialCharacters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/function?function_name=Tool+%26+QA&project_id=project%2F123&slug=tool%2Falpha"
		if got := r.URL.RequestURI(); got != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, got)
		}

		resp := ListFunctionsResponse{Functions: []Function{{ID: "function-1"}}}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("sk-test", server.URL, "org-test")
	c.httpClient = server.Client()

	_, err := c.ListFunctions(context.Background(), &ListFunctionsOptions{
		ProjectID:    "project/123",
		FunctionName: "Tool & QA",
		Slug:         "tool/alpha",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsFunctionNotFound(t *testing.T) {
	testCases := map[string]struct {
		err  error
		want bool
	}{
		"matches API 400 with expected message": {
			err:  &APIError{StatusCode: 400, Message: "Function does not exist or you do not have access"},
			want: true,
		},
		"does not match API 404": {
			err:  &APIError{StatusCode: 404, Message: "Function does not exist or you do not have access"},
			want: false,
		},
		"does not match different API 400 message": {
			err:  &APIError{StatusCode: 400, Message: "invalid request"},
			want: false,
		},
		"does not match non-api errors": {
			err:  errors.New("boom"),
			want: false,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := IsFunctionNotFound(tc.err); got != tc.want {
				t.Fatalf("IsFunctionNotFound() = %t, want %t", got, tc.want)
			}
		})
	}
}
