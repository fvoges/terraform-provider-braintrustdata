package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetScore(t *testing.T) {
	position := "0|hzzzz:"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/project_score/score-123" {
			t.Errorf("expected path /v1/project_score/score-123, got %s", r.URL.Path)
		}

		resp := ProjectScore{
			ID:          "score-123",
			Name:        "quality",
			ProjectID:   "proj-123",
			UserID:      "user-123",
			ScoreType:   "categorical",
			Description: "Quality score",
			Created:     "2026-03-01T00:00:00Z",
			Position:    &position,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	score, err := client.GetScore(context.Background(), "score-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score.ID != "score-123" {
		t.Errorf("expected id score-123, got %s", score.ID)
	}
	if score.Name != "quality" {
		t.Errorf("expected name quality, got %s", score.Name)
	}
	if score.Position == nil || *score.Position != "0|hzzzz:" {
		t.Errorf("expected position 0|hzzzz:, got %v", score.Position)
	}
}

func TestGetScore_EmptyID(t *testing.T) {
	client := NewClient("sk-test", "https://api.braintrust.dev", "org-test")

	_, err := client.GetScore(context.Background(), "")
	if !errors.Is(err, ErrEmptyScoreID) {
		t.Fatalf("expected ErrEmptyScoreID, got %v", err)
	}
}

func TestListScores_WithOptions(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/project_score" {
			t.Errorf("expected path /v1/project_score, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("limit"); got != "10" {
			t.Errorf("expected limit 10, got %q", got)
		}
		if got := query.Get("starting_after"); got != "score-next" {
			t.Errorf("expected starting_after score-next, got %q", got)
		}
		if got := query.Get("ending_before"); got != "score-prev" {
			t.Errorf("expected ending_before score-prev, got %q", got)
		}
		if got := query.Get("project_score_name"); got != "quality" {
			t.Errorf("expected project_score_name quality, got %q", got)
		}
		if got := query.Get("project_name"); got != "example-project" {
			t.Errorf("expected project_name example-project, got %q", got)
		}
		if got := query.Get("project_id"); got != "proj-123" {
			t.Errorf("expected project_id proj-123, got %q", got)
		}
		if got := query.Get("score_type"); got != "categorical" {
			t.Errorf("expected score_type categorical, got %q", got)
		}
		if got := query.Get("org_name"); got != "example-org" {
			t.Errorf("expected org_name example-org, got %q", got)
		}
		if got := query["ids"]; !reflect.DeepEqual(got, []string{"score-1", "score-2"}) {
			t.Errorf("expected ids [score-1 score-2], got %v", got)
		}

		resp := ListScoresResponse{
			Objects: []ProjectScore{{
				ID:        "score-1",
				Name:      "quality",
				ProjectID: "proj-123",
				UserID:    "user-1",
			}},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	result, err := client.ListScores(context.Background(), &ListScoresOptions{
		Limit:         10,
		StartingAfter: "score-next",
		EndingBefore:  "score-prev",
		IDs:           []string{"score-1", "score-2"},
		ScoreName:     "quality",
		ProjectName:   "example-project",
		ProjectID:     "proj-123",
		ScoreType:     "categorical",
		OrgName:       "example-org",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Objects) != 1 {
		t.Fatalf("expected 1 score, got %d", len(result.Objects))
	}
	if result.Objects[0].ID != "score-1" {
		t.Errorf("expected score id score-1, got %s", result.Objects[0].ID)
	}
}

func TestListScores_SpecialCharacters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/project_score?project_name=Project+%26+Co&project_score_name=v1%2Fbeta"
		if got := r.URL.RequestURI(); got != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, got)
		}

		resp := ListScoresResponse{Objects: []ProjectScore{{ID: "score-1"}}}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", server.URL, "org-test")
	client.httpClient = server.Client()

	_, err := client.ListScores(context.Background(), &ListScoresOptions{
		ProjectName: "Project & Co",
		ScoreName:   "v1/beta",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
