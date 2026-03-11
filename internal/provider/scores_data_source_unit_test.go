package provider

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildListScoresOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantErrLike string
		want        client.ListScoresOptions
		model       ScoresDataSourceModel
	}{
		"builds all supported API-native filters": {
			model: ScoresDataSourceModel{
				FilterIDs:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("score-1"), types.StringValue("score-2")}),
				OrgName:       types.StringValue("example-org"),
				ProjectID:     types.StringValue("proj-1"),
				ProjectName:   types.StringValue("example-project"),
				ScoreName:     types.StringValue("quality"),
				ScoreType:     types.StringValue("categorical"),
				StartingAfter: types.StringValue("score-10"),
				Limit:         types.Int64Value(10),
			},
			want: client.ListScoresOptions{
				IDs:           []string{"score-1", "score-2"},
				OrgName:       "example-org",
				ProjectID:     "proj-1",
				ProjectName:   "example-project",
				ScoreName:     "quality",
				ScoreType:     "categorical",
				StartingAfter: "score-10",
				Limit:         10,
			},
		},
		"rejects conflicting pagination": {
			model: ScoresDataSourceModel{
				StartingAfter: types.StringValue("score-1"),
				EndingBefore:  types.StringValue("score-2"),
			},
			wantErrLike: "Cannot specify both 'starting_after' and 'ending_before'",
		},
		"rejects zero limit": {
			model: ScoresDataSourceModel{
				Limit: types.Int64Value(0),
			},
			wantErrLike: "'limit' must be greater than or equal to 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts, diags := buildListScoresOptions(context.Background(), tc.model)
			if tc.wantErrLike != "" {
				if !diags.HasError() {
					t.Fatalf("expected diagnostic containing %q, got none", tc.wantErrLike)
				}
				found := false
				for _, diag := range diags {
					if strings.Contains(diag.Detail(), tc.wantErrLike) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected diagnostic containing %q, got %v", tc.wantErrLike, diags)
				}

				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if opts == nil {
				t.Fatalf("expected options, got nil")
			}

			if !reflect.DeepEqual(opts.IDs, tc.want.IDs) ||
				opts.OrgName != tc.want.OrgName ||
				opts.ProjectID != tc.want.ProjectID ||
				opts.ProjectName != tc.want.ProjectName ||
				opts.ScoreName != tc.want.ScoreName ||
				opts.ScoreType != tc.want.ScoreType ||
				opts.StartingAfter != tc.want.StartingAfter ||
				opts.EndingBefore != tc.want.EndingBefore ||
				opts.Limit != tc.want.Limit {
				t.Fatalf("options mismatch: got=%+v want=%+v", *opts, tc.want)
			}
		})
	}
}

func TestScoresDataSourceScoreFromScore(t *testing.T) {
	t.Parallel()

	position := "0|hzzzz:"
	score := &client.ProjectScore{
		ID:          "score-1",
		ProjectID:   "project-1",
		UserID:      "user-1",
		Created:     "2026-03-02T00:00:00Z",
		Name:        "quality",
		Description: "Quality score",
		ScoreType:   "categorical",
		Categories:  []string{"good", "bad"},
		Config:      map[string]interface{}{"max": float64(5)},
		Position:    &position,
	}

	model, diags := scoresDataSourceScoreFromScore(score)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
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

func TestScoresDataSourceScoreFromScore_NullPosition(t *testing.T) {
	t.Parallel()

	score := &client.ProjectScore{
		ID:        "score-2",
		ProjectID: "project-2",
		Name:      "latency",
	}

	model, diags := scoresDataSourceScoreFromScore(score)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.Position.IsNull() {
		t.Fatalf("expected position to be null")
	}
}

func TestProviderDataSourcesIncludeScorePair(t *testing.T) {
	t.Parallel()

	p, ok := New("test")().(*BraintrustProvider)
	if !ok {
		t.Fatalf("expected *BraintrustProvider")
	}

	dataSourceFactories := p.DataSources(context.Background())
	dataSourceNames := make(map[string]struct{}, len(dataSourceFactories))

	for _, factory := range dataSourceFactories {
		ds := factory()
		resp := &datasource.MetadataResponse{}
		ds.Metadata(context.Background(), datasource.MetadataRequest{
			ProviderTypeName: "braintrustdata",
		}, resp)

		dataSourceNames[resp.TypeName] = struct{}{}
	}

	if _, ok := dataSourceNames["braintrustdata_score"]; !ok {
		t.Fatalf("expected braintrustdata_score to be registered")
	}
	if _, ok := dataSourceNames["braintrustdata_scores"]; !ok {
		t.Fatalf("expected braintrustdata_scores to be registered")
	}
}
