package provider

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
)

func TestSelectSingleScoreByName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrType error
		scoreName   string
		wantID      string
		scores      []client.ProjectScore
	}{
		"finds exact score": {
			scores: []client.ProjectScore{
				{ID: "score-a", Name: "other"},
				{ID: "score-b", Name: "target"},
			},
			scoreName: "target",
			wantID:    "score-b",
		},
		"returns not found when no exact match": {
			scores: []client.ProjectScore{
				{ID: "score-a", Name: "other"},
			},
			scoreName:   "target",
			wantErrType: errScoreNotFoundByName,
		},
		"returns multiple when exact matches are ambiguous": {
			scores: []client.ProjectScore{
				{ID: "score-a", Name: "target"},
				{ID: "score-b", Name: "target"},
			},
			scoreName:   "target",
			wantErrType: errMultipleScoresFoundByName,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			score, err := selectSingleScoreByName(tc.scores, tc.scoreName)
			if tc.wantErrType != nil {
				if !errors.Is(err, tc.wantErrType) {
					t.Fatalf("expected error %v, got %v", tc.wantErrType, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if score == nil {
				t.Fatalf("expected score, got nil")
			}
			if score.ID != tc.wantID {
				t.Fatalf("expected score ID %q, got %q", tc.wantID, score.ID)
			}
		})
	}
}

func TestPopulateScoreDataSourceModel(t *testing.T) {
	t.Parallel()

	position := "0|hzzzz:"
	model := ScoreDataSourceModel{}
	score := &client.ProjectScore{
		ID:          "score-1",
		Name:        "quality",
		ProjectID:   "project-1",
		UserID:      "user-1",
		Created:     "2026-03-02T00:00:00Z",
		Description: "Quality score",
		ScoreType:   "categorical",
		Categories:  []string{"good", "bad"},
		Config:      map[string]interface{}{"max": float64(5)},
		Position:    &position,
	}

	diags := populateScoreDataSourceModel(context.Background(), &model, score)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "score-1" {
		t.Fatalf("id mismatch: got=%q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "quality" {
		t.Fatalf("name mismatch: got=%q", model.Name.ValueString())
	}
	if model.ProjectID.ValueString() != "project-1" {
		t.Fatalf("project_id mismatch: got=%q", model.ProjectID.ValueString())
	}
	if model.UserID.ValueString() != "user-1" {
		t.Fatalf("user_id mismatch: got=%q", model.UserID.ValueString())
	}
	if model.Created.ValueString() != "2026-03-02T00:00:00Z" {
		t.Fatalf("created mismatch: got=%q", model.Created.ValueString())
	}
	if model.Description.ValueString() != "Quality score" {
		t.Fatalf("description mismatch: got=%q", model.Description.ValueString())
	}
	if model.ScoreType.ValueString() != "categorical" {
		t.Fatalf("score_type mismatch: got=%q", model.ScoreType.ValueString())
	}
	if model.Position.ValueString() != "0|hzzzz:" {
		t.Fatalf("position mismatch: got=%q", model.Position.ValueString())
	}

	var categories []string
	if err := json.Unmarshal([]byte(model.Categories.ValueString()), &categories); err != nil {
		t.Fatalf("categories should be valid JSON: %v", err)
	}
	if len(categories) != 2 || categories[0] != "good" || categories[1] != "bad" {
		t.Fatalf("unexpected categories: %v", categories)
	}

	var config map[string]float64
	if err := json.Unmarshal([]byte(model.Config.ValueString()), &config); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}
	if config["max"] != 5 {
		t.Fatalf("config mismatch: %v", config)
	}
}

func TestPopulateScoreDataSourceModel_Nullables(t *testing.T) {
	t.Parallel()

	model := ScoreDataSourceModel{}
	score := &client.ProjectScore{
		ID:        "score-2",
		Name:      "latency",
		ProjectID: "project-2",
	}

	diags := populateScoreDataSourceModel(context.Background(), &model, score)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.UserID.IsNull() {
		t.Fatalf("expected user_id to be null")
	}
	if !model.Created.IsNull() {
		t.Fatalf("expected created to be null")
	}
	if !model.Description.IsNull() {
		t.Fatalf("expected description to be null")
	}
	if !model.ScoreType.IsNull() {
		t.Fatalf("expected score_type to be null")
	}
	if !model.Categories.IsNull() {
		t.Fatalf("expected categories to be null")
	}
	if !model.Config.IsNull() {
		t.Fatalf("expected config to be null")
	}
	if !model.Position.IsNull() {
		t.Fatalf("expected position to be null")
	}
}
